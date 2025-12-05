package nftables

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/netip"

	"github.com/google/nftables"
)

func Connect() (*nftables.Conn, error) {
	nft, err := nftables.New()
	if err != nil {
		slog.Error("Failure Connect() => New()", "error", err)
		return nil, err
	}

	return nft, nil
}

type nftError struct {
	Op  string
	Err error
}

func (e *nftError) Unwrap() error {
	return e.Err
}

func (e *nftError) Error() string {
	return fmt.Sprintf("%s => %s", e.Op, e.Err)
}

func newNftError(op string, inner error) error {
	e := &nftError{
		Op:  op,
		Err: inner,
	}

	slog.Debug(e.Error())

	return e
}

var (
	ErrUnknownFamily = errors.New("Unknown family")
)

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

	var (
		out  []string
		last nftables.SetElement
	)

	NullIPv4 := []byte{0, 0, 0, 0}
	NullIPv6 := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	//for _, e := range elements {
	// iterate in reverse: https://github.com/google/nftables/issues/320
	for i := len(elements) - 1; i >= 0; i-- {
		e := elements[i]
		switch set.KeyType.Name {
		case "ipv4_addr", "ipv6_addr":
			// https://github.com/google/nftables/issues/346

			if bytes.Compare(e.Key, NullIPv4) == 0 || bytes.Compare(e.Key, NullIPv6) == 0 {
				continue
			}

			ipCurrent, _ := netip.AddrFromSlice(e.Key)
			ipLast, _ := netip.AddrFromSlice(last.Key)
			ipLastNext := ipLast.Next()

			if e.IntervalEnd && ipCurrent == ipLastNext {
				out = append(out, ipLast.String())
				continue
			}

			if e.IntervalEnd {
				maxLen := 32
				if ipLastNext.Is6() {
					if !ipCurrent.Is6() {
						continue
					}
					maxLen = 128
				}
				for l := maxLen; l >= 0; l-- {
					mask := net.CIDRMask(l, maxLen)
					na := net.IP(ipLastNext.AsSlice()).Mask(mask)
					n := net.IPNet{IP: na, Mask: mask}
					if n.Contains(net.IP(ipCurrent.AsSlice())) {
						out = append(out, fmt.Sprintf("%s/%d", na, l+1))
						break
					}
				}
				continue
			}

			last = e

		default:
			slog.Error("Unimplemented set data type", "SetName", set.Name, "KeyTypeName", set.KeyType.Name)
		}
	}

	return out, nil
}

func GetSetFlags(set *nftables.Set) (flags []string) {
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
