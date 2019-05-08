# Tangle Authentication

## What is it
Tangle authentication is a POC for decentralized authentication through the iota tangle.

Iota is a distributed ledger and cryptocurrency that doesn't use a blockchain but a structure called the tangle more info can be found [here](https://www.iota.org/).

Because iota feeless and allows 0 value messages we can use it to store identity data in a decentralized way.

## How to run it
Install the package with
```
go get github.com/pkartner/tau
```
You can find 2 example applications in the example folders. Go to the client folder in examples and run
```
go run .\main.go create -s myseed -n myname
```
Replace seed with a random string. Save it because you will need it later and there is no way to retrieve it. This will create an account on the tangle. Optionally you can also add the -n with a name, this will attach your name to your id.

Go to the server folder in examples and run
```
go run .\main.go
```
This will run a local server we will use to test out our authentication.

Go back to the client folder and run 
```
go run .\main.go request -s myseed
```
Use the same seed you used to create your account. This will make a request to your local server and authenticate you using your seed.

## How does it work
Tangle authentication lets you claim an iota address on the tangle (based on [iota-origin](https://medium.com/coinmonks/iota-origin-d26b399d3cca)), making this address your ID. You can later proof that you own this address letting you use it for identification purposes. It is also possible to store personal data with this ID.

### Creating an ID
1. Give a seed, this can be anything but it is best to use a long random string.
2. Generate a private/public key pair using this seed as an input.
3. Sign a hash of the public key.
4. Hash this signature to a 81 tryte(A-Z9) string (this is the format of an iota address) and add a 9 tryte checksum. This will be your ID.
5. Add the public key and the hash to a json message and send this to the tangle address we generated.

### Authenticating Client
1. Take your seed.
2. Generate a private/public key pair using this seed as an input.
3. Sign a hash of the public key.
4. Hash this signature to a 81 tryte(A-Z9) string (this is the format of an iota address) and add a 9 tryte checksum. This will be your ID.
5. Construct a JWT with address as the sub and the url to which we are making the request as the aud.
6. Sign the JWT using our private key.
7. Send the request with the JWT in the Authorization header.

### Authenticating Server
1. Extract the JWT from the Authorization header.
2. Extract the claims from the token.
3. Check if the requested url matches the url that is in the aud claim.
4. Find the sub claim, this is your ID and an iota address. 
5. Request all the transactions from this address from the tangle and sort these transactions from newest to oldest. 
6. Loop through these transactions and check the rest of the list. If a transaction passes all test you are done and the user is authorized. If no transaction passes all the tests than the user isn't authorized.
7. Use the public key from the transaction message to verify the JWT signature.
8. Verify the signature in the transaction message with the public key.
9. Hash this signature to a 81 tryte(A-Z9) string (this is the format of an iota address) and add a 9 tryte checksum. Match this against the sub claim to make sure they are the same.
