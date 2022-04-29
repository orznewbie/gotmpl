package dgraph

import (
	"context"
	"fmt"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/orznewbie/gotmpl/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func dgoClient() (*dgo.Dgraph, *grpc.ClientConn) {
	cc, err := grpc.Dial("192.168.30.58:29080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	dc := api.NewDgraphClient(cc)
	dg := dgo.NewDgraphClient(dc)

	return dg, cc
}

func TestDropType(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	if err := dg.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "dtdl:test:Space"}); err != nil {
		t.Fatal(err)
	}
}

func TestDropData(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	if err := dg.Alter(context.Background(), &api.Operation{
		DropOp: api.Operation_DATA,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDropPred(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	dropAttr := `dtdl:test:Room::area dtdl:test:Room::room_number`

	schema := `
	type <dtdl:test:Room> {
		dtdl:core:Metadata::etag
		dtdl:core:Metadata::extends
		dtdl:core:Metadata::revision
		dtdl:core:Metadata::version
		dtdl:core:Metadata::create_time
		dtdl:core:Metadata::update_time
		dtdl:core:Metadata::delete_time
		dtdl:test:Space::space_number
		dtdl:test:Space::capacity

		dtdl:test:Space::is_part_of
		dtdl:core:Object::displayName
		dtdl:core:Object::discription
		dtdl:core:Object::comment
	}
	`

	if err := dg.Alter(context.Background(), &api.Operation{
		Schema:   schema,
		DropAttr: dropAttr}); err != nil {
		t.Fatal(err)
	}
}

func TestAddNoconflictSchema(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	op := &api.Operation{}
	op.Schema = `
		email: string @noconflict .
	`
	ctx := context.Background()
	if err := dg.Alter(ctx, op); err != nil {
		t.Fatal(err)
	}
}

func TestQueryRDF(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	q := `
	query {
	  q(func:has(dgraph.type)) {
		uid
		expand(_all_)
	  }
	}`

	resp, err := dg.NewTxn().QueryRDF(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(resp.Rdf))
}

func TestQuerySchema(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	q := `schema(type: <dtdl:test:Room;1>) {}`
	resp, err := dg.NewTxn().Query(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(resp.Json))
}

func TestQuery(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	q := `
	query {
	  q(func:has(dgraph.type),first:2) {
		uid
	  }
	}`

	resp, err := dg.NewReadOnlyTxn().Query(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(resp.Json))
}

func TestMutate(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	quad := `
	_:user <name> "lisi" .
	_:user <age> "35" .
	_:user <dgraph.type> "User" .
	`

	tx := dg.NewTxn()
	resp, err := tx.Mutate(context.Background(), &api.Mutation{
		SetNquads: []byte(quad),
		CommitNow: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	tx.Commit(context.Background())

	log.Info(resp)

}

func TestTx(t *testing.T) {
	dg, cc := dgoClient()
	defer cc.Close()

	txA := dg.NewTxn()
	time.Sleep(time.Second*2)
	txB := dg.NewTxn()

	resp, err := txA.Mutate(context.TODO(), &api.Mutation{
		SetNquads:  []byte(`<0xc351> <name> "AAA" .`),
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("txA: ", resp.Txn)

	resp, err = txB.Mutate(context.TODO(), &api.Mutation{
		SetNquads:  []byte(`<0xc351> <name> "BBB" .`),
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("txB: ", resp.Txn)

	if err := txA.Commit(context.TODO()); err != nil {
		log.Error(err)
	}
	if err := txB.Commit(context.TODO()); err != nil {
		log.Error(err)
	}
}
