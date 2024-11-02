package main

import (
	"context"
	"flag"
	"fmt"

	hashiVault "github.com/hashicorp/vault/api"
	"github.com/jacobbrewer1/vaulty"
)

var (
	vaultAddr = flag.String("addr", "http://localhost:8200", "The address of the vault server")
	vaultUser = flag.String("user", "root", "The username to authenticate with")
	vaultPass = flag.String("pass", "root", "The password to authenticate with")
	vaultPath = flag.String("path", "secret", "The path to the secrets")
	dbHost    = flag.String("host", "localhost:3306", "The host of the database")
	dbSchema  = flag.String("schema", "test", "The schema of the database")
)

func init() {
	flag.Parse()
}

func generateConnectionStr(vs *hashiVault.Secret) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=90s&multiStatements=true&parseTime=true",
		vs.Data["username"],
		vs.Data["password"],
		*dbHost,
		*dbSchema,
	)
}

func main() {
	vc, err := vaulty.NewClient(
		vaulty.WithGeneratedVaultClient(*vaultAddr),
		vaulty.WithUserPassAuth(
			*vaultUser,
			*vaultPass,
		),
	)
	if err != nil {
		panic(fmt.Errorf("error creating vault client: %w", err))
	}

	got, err := vc.Path(*vaultPath).GetSecret(context.Background())
	if err != nil {
		panic(fmt.Errorf("error getting secret: %w", err))
	}

	fmt.Println(generateConnectionStr(got))
}
