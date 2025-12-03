package nftables

import (
	"errors"
	"log/slog"

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
		slog.Error("Failure getTable() => unknown family")
		return nil, errors.New("unknown family")
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

	var out []string

	for _, e := range elements {
		// TODO
		out = append(out, string(e.Key[:]))
	}

	return out, nil
}
