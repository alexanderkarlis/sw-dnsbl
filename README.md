# DNS Blocklist
GraphQL API built in Go to check whether or a not a specific IP address is suspected of spam against a DNS. Using DNS requests, you can send an IP address in reverse notation to determine whether it is on a block list or not. 

## Technical overview
___
The three main part of the microservice are as follows:

* MySql Database for data persistance 
* GraphQL server for incoming requests to the database
* Go programming interface to stand up GraphQL requests
* Go Tests
* Dockerfile for building application

### **Go.mod deps** ###
- **[mysql](github.com/go-sql-driver/mysql)** - MySQL database driver
- **[gqlgen](https://github.com/99designs/gqlgen)** - Generates Go template code for GraphQL servers
- **[testify](https://github.com/stretchr/testify)** - Tools for testifying that your code will behave as you intend (asserts, requires, etc..)
- **[go-sqlmock](github.com/DATA-DOG/go-sqlmock)** - mocking database calls for testing 
- **[github.com/alexanderkarlis/godnsbl](github.com/alexanderkarlis/godnsbl)** - DNS Blocklist lookup functionality (forked from github.com/mrichman/godnsbl)
- **[jwt-go](https://github.com/dgrijalva/jwt-go)** - Creating JWTs
- **[mux](github.com/gorilla/mux)** - a request router and dispatcher for matching incoming requests to their respective handler.
- **[uuid](https://github.com/google/uuid)** - Generates UUIDs
- **[gqlparser](https://github.com/vektah/gqlparser/v2)** - This is a parser for graphql

#### *GoDNSBL package*
Needed a fork because, from my understanding, the package in its current state was not returning the correct codes for some based on a binary `true`/`false`. If the result was true, the DNSBL package in this application would return the checked IP address, which is correct for most but not all (at least in the case of only `zen.spamhaus.org`). In the `Result` struct, `Code` field was added to Lookup result struct for more accurate response.

### **Program structure** ###
```
.
├── auth
│   ├── auth.go
│   └── auth_test.go
├── config
│   └── config.go
├── config.env
├── database
│   ├── database.go
│   └── database_test.go
├── dnsbl
│   ├── dnsbl.go
│   └── dnsbl_test.go
├── docker-compose-script.sh
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── gqlgen.yml
├── graph
│   ├── generated
│   │   └── generated.go
│   ├── model
│   │   └── models_gen.go
│   ├── resolver.go
│   ├── schema.graphqls
│   └── schema.resolvers.go
├── helm
│   ├── charts
│   ├── Chart.yaml
│   ├── templates
│   │   ├── deployment.yaml
│   │   ├── _helpers.tpl
│   │   ├── hpa.yaml
│   │   ├── ingress.yaml
│   │   ├── NOTES.txt
│   │   ├── serviceaccount.yaml
│   │   ├── service.yaml
│   │   └── tests
│   │       └── test-connection.yaml
│   └── values.yaml
├── logs
│   └── app.log
├── middleware
│   ├── middleware.go
│   └── middleware_test.go
├── README.md
├── scripts
│   └── ip
│       ├── init.sql
│       └── upsert.sql
├── server.go
├── start-helm.sh
└── start_server.sh
``` 
The **auth**, **config**, **database**, **dnsbl**, **graph**, **middleware** folders contain all the package code for the Go code. Below are the packages main functions and additional information therefore.

### Auth
The GraphQL API has a basic authentication layer allowing only authenticated users to use the it. Upon successful authentication, a user is granted a `bearer` token which can be used to access the API. Right now, the app only allows for one user. This is stored in the [GraphQL resolvers](graph/schema.resolvers.go).
___
- **Username** : secureworks
- **Password** : supersecret
1. `CreateJWT` --> grants a user a JSON Web Token upon successful login (**JWT**)
2. `ValidateToken` --> validates a JWT
3. GraphQL **`authenticate`** mutation

### Config
Configurations are taken as environment variables from `config.env` file.
___
1. `GetConfig` --> Go function that gets all the listed Environment Variables and passes them to the main app server. See [config.env](./config.env) for the configuration possibilities. *NOTE: env some variables have default values.* 

### Database
MySQL. To get a glimpse of the overall db schema, see the [GraphQL section](#schema)
___
1. `NewDb` --> Function for creating a new database instance
2. `UpsertRecord` --> takes in a struct from the generated GraphQL model and inserts if the record doesn't exist, otherwise, the record is updated
3. `QueryRecord` --> takes in a string of IP Address to query

&emsp;[to database section](#database)

### Dnsbl
This is where the main DNS Blocklist lookup happens. This also contains the `consumer`, which houses the `queue` of workers. The `queue` is a list of IP addresses used to do a Blocklist check by proxy of `godnsbl.Lookup`.
___
1. `NewConsumer` --> Returns a new consumer defined by worker poolsize and the DNS Blocklist from the config env vars and database instance. Kicks off go-routine worker so it can `listen` for the changes to the jobs channel.
2. `Queue` --> Queues up a an array of ip addresses to send the the jobs channel. Main function for the `enqueue` GraphQL mutation.
3. `worker` --> a worker triggered by a slice of IPs sent to the jobs channel. Iterates through the array of IPs running the `godnsbl.Lookup`.

*note on [github.com/alexanderkarlis/godnsbl](github.com/alexanderkarlis/godnsbl)*; the lookup function could possibly return multiple `return codes`. Thus we have to account for that by taking the first one in the list. This is best explained in `server_test.go` unit tests for a few of the queries; [see](server_test.go) line #

### GraphQL
The config.env file holds the default port for the api server at `8080`, which can be changed.
_____
The two main end points:

- `enqueue` - mutation to kick off a background job and stores it in
the database for each IP passed in. If the lookup has already happened, this will queue it up again and update the `response`​ and `updated_at`​ fields in the db
- `getIPDetails` - query for obtaining blocklist details for a single IP address. The response code field is designated from the values of [zen.spamhaus.org](https://www.spamhaus.org/faq/section/DNSBL%20Usage#200)


<a id="schema"></a>Schema 

See [schema.graphqls](graph/ip/init.sql) for more details.

### Logging
Logging to a file is set to an environment variable in the `config.env` file. 

## Local Development
### <a id="database"></a>Development with a local database
As mentioned, the chosen database was MySql. In order to do development with a local MySql instance, a few things need to be configured. The API uses a connection string with a username, password, host name, host port (default: 3306), and database name all from the `config.env` file. 

#### Defaults from the config.env file
- username: `mysql_admin` <br/>
- password: `password` <br/>
- host_name: `localhost` || `db` *(if running in docker-compose; reference to the `db` container)* <br/>
- host_port: `3306` <br/>
- database_name: `sw_dnsbl`

1. Configure DB with appropriate username and password
2. Create database in accoridance with `config.env`
1. Run the [init.sql](scripts/ip/init.sql) script to create the table.
2. This script sets the environment variables from the `config.env` file and starts running the server locally ```./start_local_server.sh```

## Docker
This package can be run locally with the following supplied script:

    ./run-docker.sh [port]
   
where `port` is an optional argument to run the docker container on. Default **8080**.

## Docker-Compose
This package can be run locally with the following supplied script:
```sh
> rm -rf ./tmp/  # remove db data
> docker-compose build
> docker-compose up
```   
## Helm
Install a helm chart:
```sh
helm install swdnsbl ./helm
```

## TODO's
- [ ] Add some role based authentication the more the api grows in complexity
- [ ] Right now, the data model only allows for one Domain to be checked (`zen.spamhaus.org`). Update data model to reflect multiple Domains and IP checks
- [ ] Add more specific logging
- [ ] Add more queries/ mutations for more insight into the problem
