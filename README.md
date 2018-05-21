# Graphql Go Auth

Enables you to easily handle authentication with [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go). This project currently provides Oauth2 authentication with google.

## Example use

```
# 1- go to example
cd example

# 2- use the script to start and watch
./start.sh

# 3- edit example/_config/global.json
# 4- visit localhost:8080
```
Then you can enter this in GraphiQL. Make sure your google client id and key are correct in global.json.
```graphql
mutation {
  triggerOauth(input: {
    username: "bob", password: "123"
  })
}
```
