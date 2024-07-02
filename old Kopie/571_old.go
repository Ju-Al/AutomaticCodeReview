package sda

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"io/ioutil"

	"github.com/dedis/cothority/log"
)

func TestSimulationBF(t *testing.T) {
	log.AfterTest(t)
	sc, _, err := createBFTree(7, 2)
	if err != nil {
		t.Fatal(err)
	}
	addresses := []string{
		"127.0.0.1:2000",
		"127.0.0.1:2001",
		"127.0.0.1:2002",
		"127.0.0.1:2003",
		"127.0.0.1:2004",
		"127.0.0.1:2005",
		"127.0.0.1:2006",
	}
	for i, a := range sc.Roster.List {
		if a.Address.NetworkAddress() != addresses[i] {
			t.Fatal("Address", a.Address.NetworkAddress(), "should be", addresses[i])
		}
	}
	if !sc.Tree.IsBinary(sc.Tree.Root) {
		t.Fatal("Created tree is not binary")
	}

	sc, _, err = createBFTree(13, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Tree.Root.Children) != 3 {
		t.Fatal("Branching-factor 3 tree has not 3 children")
	}
	if !sc.Tree.IsNary(sc.Tree.Root, 3) {
		t.Fatal("Created tree is not binary")
	}
}

func TestBigTree(t *testing.T) {
	t.Skip()
	log.AfterTest(t)
	if testing.Short() {
		t.Skip()
	}
	for i := uint(12); i < 15; i++ {
		_, _, err := createBFTree(1<<i-1, 2)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSimulationLoadSave(t *testing.T) {
	log.AfterTest(t)
	sc, _, err := createBFTree(7, 2)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir("", "example")
	log.ErrFatal(err)
	defer os.RemoveAll(dir)
	sc.Save(dir)
	sc2, err := LoadSimulationConfig(dir, "127.0.0.1:2000")
	if err != nil {
		t.Fatal(err)
	}
	defer closeAll(sc2)
	if sc2[0].Tree.ID != sc.Tree.ID {
		t.Fatal("Tree-id is not correct")
	}
}

func TestSimulationMultipleInstances(t *testing.T) {
	log.AfterTest(t)
	sc, _, err := createBFTree(7, 2)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir("", "example")
	log.ErrFatal(err)
	defer os.RemoveAll(dir)
	sc.Save(dir)
	sc2, err := LoadSimulationConfig(dir, "local1")
	if err != nil {
		t.Fatal(err)
	}
	defer closeAll(sc2)
	if len(sc2) != 7 {
		t.Fatal("We should have 4 local1-hosts but have", len(sc2))
	}
	if sc2[0].Host.ServerIdentity.ID == sc2[1].Host.ServerIdentity.ID {
		t.Fatal("Hosts are not copies")
	}
}

func closeAll(scs []*SimulationConfig) {
	for _, s := range scs {
		s.Host.Close()
	}
}

func createBFTree(hosts, bf int) (*SimulationConfig, *SimulationBFTree, error) {
	sc := &SimulationConfig{}
	sb := &SimulationBFTree{
		Hosts: hosts,
		BF:    bf,
	}
	sb.CreateRoster(sc, []string{"127.0.0.1"}, 2000)
	if len(sc.Roster.List) != hosts {
		return nil, nil, errors.New("Didn't get correct number of entities")
	}
	err := sb.CreateTree(sc)
	if err != nil {
		return nil, nil, err
	}
	if !sc.Tree.IsNary(sc.Tree.Root, bf) {
		return nil, nil, errors.New("Tree isn't " + strconv.Itoa(bf) + "-ary")
	}

	return sc, sb, nil
}
