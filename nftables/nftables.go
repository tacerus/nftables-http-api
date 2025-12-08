package nftables

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/netip"
	"slices"

	"github.com/google/nftables"
)

var (
	SingleAddrMaskIPv4 = bytes.Repeat([]byte{255}, 4)
	SingleAddrMaskIPv6 = bytes.Repeat([]byte{255}, 16)
)

func Connect() (*nftables.Conn, error) {
	nft, err := nftables.New()
	if err != nil {
		slog.Error("Failure Connect() => New()", "error", err)
		return nil, err
	}

	return nft, nil
}

func getFamily(familyName string) (family nftables.TableFamily) {
	switch familyName {
	case "inet":
		family = nftables.TableFamilyINet
	case "ipv4":
		family = nftables.TableFamilyIPv4
	case "ipv6":
		family = nftables.TableFamilyIPv6
	default:
		family = nftables.TableFamilyUnspecified
	}

	return family
}

// TODO: replace with ListTableOfFamily()
func GetTable(nft *nftables.Conn, familyName string, tableName string) (*nftables.Table, error) {
	tables, err := nft.ListTables()
	if err != nil {
		slog.Error("Failure getTable() => ListTables()", "error", err)
		return nil, err
	}

	family := getFamily(familyName)
	if family == nftables.TableFamilyUnspecified {
		return nil, newNftError("getTable()", ErrUnknownFamily)
	}

	for _, t := range tables {
		if t.Family == family && t.Name == tableName {
			return t, nil
		}
	}

	return nil, nil
}

func GetSet(nft *nftables.Conn, table *nftables.Table, setName string) (*nftables.Set, error) {
	set, err := nft.GetSetByName(table, setName)
	if err != nil {
		slog.Error("Failure getSet() => GetSetByName()", "error", err)
		return nil, err
	}

	if set == nil {
		return nil, nil
	}

	return set, nil
}

func GetSetElements(nft *nftables.Conn, set *nftables.Set) ([]string, error) {
	elements, err := nft.GetSetElements(set)
	if err != nil {
		slog.Error("Failure GetSetElements() => GetSetElements()", "error", err)
		return nil, err
	}

	out := []string{}
	var start []byte

	// https://github.com/google/nftables/issues/320
	slices.Reverse(elements)

	for i, e := range elements {
		eaddr, _ := netip.AddrFromSlice(e.Key)
		slog.Debug("parsing element", "e", e, "addr", eaddr)
		switch set.KeyType.Name {
		case "ipv4_addr", "ipv6_addr":
			if i == 0 && e.IntervalEnd {
				slog.Debug("skipping first")
				continue
			}

			if !e.IntervalEnd {
				start = e.Key
				continue
			}

			if e.IntervalEnd {
				ip1, _ := netip.AddrFromSlice(start)
				ip2, _ := netip.AddrFromSlice(e.Key)
				slog.Debug("constructing net from interval range", "first", ip1, "last", ip2)
				net, ok, err := nftables.NetFromInterval(start, e.Key)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if ok {
					if ip1.Is4() && bytes.Equal(net.Mask, SingleAddrMaskIPv4) || ip1.Is6() && bytes.Equal(net.Mask, SingleAddrMaskIPv6) {
						out = append(out, net.IP.String())
					} else {
						out = append(out, net.String())
					}
				} else {
					out = append(out, fmt.Sprintf("%s-%s", ip1, ip2.Prev()))
				}
			}

			// TODO: handle IntervalOpen?

		default:
			slog.Error("Unimplemented set data type", "SetName", set.Name, "KeyTypeName", set.KeyType.Name)
			break
		}
	}

	slices.Sort(out)

	return out, nil
}

func GetSetFlags(set *nftables.Set) []string {
	flags := []string{}

	if set.Constant {
		flags = append(flags, "constant")
	}
	if set.Dynamic {
		flags = append(flags, "dynamic")
	}
	if set.Interval {
		flags = append(flags, "interval")
	}
	if set.HasTimeout {
		flags = append(flags, "timeout")
	}

	return flags
}
