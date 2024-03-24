package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	_ "github.com/lib/pq"
)

type user struct {
	Age                          int
	First_name, Last_name, Email string
}

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	host   = "postgre-postgresql.postgre.svc.cluster.local"
	port   = 5432
	dbname = "postgres"
)

const b = `
<!DOCTYPE html>
<html>
<head>
	<title>Users</title>
</head>
<body>
	<style type="text/css">
	.tg  {border-collapse:collapse;border-spacing:0;}
	.tg td{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
	overflow:hidden;padding:10px 5px;word-break:normal;}
	.tg th{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
	font-weight:normal;overflow:hidden;padding:10px 5px;word-break:normal;}
	.tg .tg-0lax{text-align:left;vertical-align:top}
	</style>
	<table class="tg">
	<thead>
	<tr>
		<th class="tg-0lax">First Name</th>
		<th class="tg-0lax">Last Name</th>
		<th class="tg-0lax">Age</th>
		<th class="tg-0lax">Email</th>
	</tr>
	</thead>
	<tbody>
	{{ range . }}
		<tr>
			<td class="tg-0lax">{{ .First_name }}</td>
			<td class="tg-0lax">{{ .Last_name }}</td>
			<td class="tg-0lax">{{ .Age }}</td>
			<td class="tg-0lax">{{ .Email }}</td>
		</tr>
	{{ end }}
	</tbody>
</body>
</html>`

const errorPage = `
<!DOCTYPE html>
<html>
<head>
	<title>Users</title>
</head>
<body>
	{{ . }}
	<br>
	<a href="/renew">Renew DB Connection</a>
</body>
</html>`

var (
	tmpl, tmplErr *template.Template
	db            *sql.DB
	err           error
)

func main() {
	if err = renewConnection(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl, err = template.New("table").Parse(b)
	if err != nil {
		log.Fatalf("could not parse main template: %v", err)
	}

	tmplErr, err = template.New("errorpage").Parse(errorPage)
	if err != nil {
		log.Fatal("could not parse error template: ", err)
	}

	log.Print("now serving webpage")
	http.HandleFunc("/", genPage)
	http.HandleFunc("/renew", renewHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func renewConnection() error {
	dbCreds, err := getDatabaseCredentials()
	if err != nil {
		return fmt.Errorf("error while obtaining database credentials from vault: %w", err)
	}

	if err = connectDB(dbCreds); err != nil {
		return fmt.Errorf("error connecting to DB: %w", err)
	}
	return nil
}

func getDatabaseCredentials() (DatabaseCredentials, error) {
	config := vault.DefaultConfig()
	config.Address = "REPLACE_VAULT_ADDR"
	// config.Address = "https://vault.vault.svc.cluster.local"
	config.ConfigureTLS(&vault.TLSConfig{Insecure: true})

	client, err := vault.NewClient(config)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	roleID := os.Getenv("APPROLE_ROLE_ID")
	if roleID == "" {
		return DatabaseCredentials{}, fmt.Errorf("no role ID was provided in APPROLE_ROLE_ID environment variable")
	}

	secretID := &approle.SecretID{
		FromEnv: "APPROLE_SECRET_ID",
	}

	appRoleAuth, err := approle.NewAppRoleAuth(
		roleID,
		secretID,
	)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return DatabaseCredentials{}, fmt.Errorf("no auth info was returned after login")
	}

	lease, err := client.Logical().ReadWithContext(context.Background(), "database/creds/readonly")
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to read secet: %w", err)
	}

	b, err := json.Marshal(lease.Data)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("malformed credentials returned from vault: %w", err)
	}

	var credentials DatabaseCredentials
	if err := json.Unmarshal(b, &credentials); err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to unmarshal credentials: %w", err)
	}

	log.Println("successfully obtained temporary database credentials from vault")
	log.Println("username is:", credentials.Username)
	return credentials, nil
}

func connectDB(dbCreds DatabaseCredentials) error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, dbCreds.Username, dbCreds.Password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	log.Print("successfully connected!")
	return nil
}

func renewHandler(w http.ResponseWriter, r *http.Request) {
	if err = renewConnection(); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/", http.StatusContinue)
}

func genPage(w http.ResponseWriter, _ *http.Request) {
	log.Print("request received, getting data from DB")
	rows, err := db.Query("SELECT age, first_name, last_name, email FROM users")
	if err != nil {
		message := fmt.Sprint("error obtaining data from DB: ", err)
		log.Print(message)
		err = tmplErr.Execute(w, message)
		if err != nil {
			log.Fatal("error executing error template: ", err)
		}
		return
	}
	defer rows.Close()

	var users []user
	for rows.Next() {
		var usr user
		if err := rows.Scan(&usr.Age, &usr.First_name, &usr.Last_name, &usr.Email); err != nil {
			message := fmt.Sprint("error parsing query result: ", err)
			log.Print(message)
			w.Write([]byte(message))
			return
		}
		users = append(users, usr)
	}
	if err = rows.Err(); err != nil {
		message := fmt.Sprint("error parsing query result: ", err)
		log.Print(message)
		w.Write([]byte(message))
		return
	}

	err = tmpl.Execute(w, users)
	if err != nil {
		log.Fatal("error executing template: ", err)
	}
}
