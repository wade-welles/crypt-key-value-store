package assn1

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	/*"fmt"*/
	/*"time"*/
	// You neet to add with
	// go get github.com/fenilfadadu/CS628-assn1/userlib
	"github.com/fenilfadadu/CS628-assn1/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging etc
	"encoding/hex"

	// UUIDs are generated right based on the crypto RNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys
	"strings"

	// Want to import errors
	"errors"
)

// Helper function to get hex encoded Argon2 hash -- Outputs 32 bytes
func Argon2Hash(toHash string) string {
	return hex.EncodeToString(userlib.Argon2Key([]byte(toHash), nil, uint32(userlib.HashSize)))
}

// Helper function to get hex encoded Argon2 hash of password -- Outputs 64 bytes
func Argon2PasswordHash(password string) string {
	return hex.EncodeToString(userlib.Argon2Key([]byte(password), nil, 2*uint32(userlib.HashSize)))
}

func MetadataHMAC(metadata MetaData, SymmetricKey []byte) ([]byte, error) {
	macInit := userlib.NewHMAC(SymmetricKey)

	macInit.Write([]byte(metadata.Owner))
	bytes, err := json.Marshal(metadata.FilenameMap)
	if err != nil {
		return nil, err
	}
	macInit.Write(bytes)
	macInit.Write([]byte(metadata.GenesisBlock))
	macInit.Write([]byte(metadata.GenesisUUIDNonce.String()))
	macInit.Write([]byte(metadata.LastBlock))
	macInit.Write([]byte(metadata.LastUUIDNonce.String()))
	return macInit.Sum(nil), nil
}

func BlockHMAC(block Block, SymmetricKey []byte) ([]byte, error) {
	macInit := userlib.NewHMAC(SymmetricKey)

	macInit.Write([]byte(block.Owner))
	macInit.Write(block.Content)
	macInit.Write([]byte(block.PrevBlockHash))
	return macInit.Sum(nil), nil
}

var fileBlocksString string = Argon2Hash("FileBlocksString")
var metaDataString string = Argon2Hash("MetaDataString")
var userDataString string = Argon2Hash("UserDataString")

// This serves two purposes: It shows you some useful primitives and
// it suppresses warnings for items not being imported
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var key *userlib.PrivateKey
	key, _ = userlib.GenerateRSAKey()
	userlib.DebugMsg("Key is %v", key)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// The structure definition for a user record
/*type File struct {
*  []byte content
*  []byte hmac
*}*/
type User struct {
	/*Username need not be encrypted with symmetric key*/
	Username     string
	SymmetricKey []byte                    // Argon2(password), given, password has high entropy
	PrivateKey   userlib.PrivateKey        // Encrypted with the Symmetric Key
	FileKeys     map[string]FileSharingKey // Indexed by filename to FileSharingKey
	HMAC         []byte                    // H(username + SymmetricKey + PrivateKey + FileKeys)
}
type FileSharingKey string // HashValue of (Owner.SymmetricKey + uuid as salt)
/*type Data struct {
 *  UserData     map[string]User
 *  FileBlocks   map[string]Block
 *  FileMetadata map[string]MetaData
 *}*/
type MetaData struct {
	Owner            string
	LastEditBy       string            // hash(LastEditByUserName)
	FilenameMap      map[string][]byte // Map from hash(username) to encrypted filename for that user (encrypted with symmetric key of that user)
	GenesisBlock     string            // HashValue(Owner + FilenameMap[Owner] + uuid nonce)
	GenesisUUIDNonce uuid.UUID
	LastBlock        string // HashValue(LastEditBy + FilenameMap[LastEditBy] + uuid nonce)
	LastUUIDNonce    uuid.UUID
	HMAC             []byte // HMAC(key = FileSharingKey, Data = Owner, LastEditBy, LastEditTime, GenesisBlock, GenesisBlockNonce, LastUUIDNonce, LastBlock)
}

type Block struct {
	Owner         string
	Content       []byte
	PrevBlockHash string
	HMAC          []byte
}
type temporaryBlock struct { //is it of any use?
	Owner         string
	Content       []byte
	PrevBlockHash string
	HMAC          []byte
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password

func InitUser(username string, password string) (userdataptr *User, err error) {

	_, ok := userlib.DatastoreGet(userDataString)
	if !ok {
		newUserMap := make(map[string]User)
		bytes, err := json.Marshal(newUserMap)
		if err != nil {
			return nil, err
		}
		userlib.DatastoreSet(userDataString, bytes)
	}
	val, ok := userlib.DatastoreGet(userDataString)
	var userDataMap map[string]User
	json.Unmarshal(val, &userDataMap)

	hashedUsername := Argon2Hash(username)

	key, err := userlib.GenerateRSAKey()
	if err != nil {
		return nil, err
	}

	// Set the Public Key in Keystore
	pubkey := key.PublicKey
	userlib.KeystoreSet(username, pubkey)

	// Populate User struct
	var userdata User
	userdata.Username = username
	userdata.SymmetricKey = []byte(Argon2PasswordHash(password))
	userdata.PrivateKey = *key
	userdata.FileKeys = make(map[string]FileSharingKey)

	macInit := userlib.NewHMAC(userdata.SymmetricKey)

	macInit.Write([]byte(userdata.Username))
	var bytes []byte
	bytes, err = json.Marshal(userdata.PrivateKey)
	if err != nil {
		return nil, err
	}
	macInit.Write(bytes)
	bytes, err = json.Marshal(userdata.FileKeys)
	if err != nil {
		return nil, err
	}
	userdata.HMAC = macInit.Sum(nil)

	// To-do CFB encryption using Symmetric Key
	userDataMap[hashedUsername] = userdata
	bytes, err = json.Marshal(userDataMap)
	if err != nil {
		return nil, err
	}
	userlib.DatastoreSet(userDataString, bytes)
	return &userdata, err
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.

/*
*		TO-DO: check for encryption, make all the errors same
 */
func GetUser(username string, password string) (userdataptr *User, err error) {
	// val contains the byte slice for the whole userDataMap
	val, ok := userlib.DatastoreGet(userDataString)
	//check if the hashedusername exists
	if !ok {
		err := errors.New("[GetUser]: userDataString wasn't indexed in Datastore.")
		return nil, err
	}
	var userDataMap map[string]User
	var userdata User
	json.Unmarshal(val, &userDataMap)

	//check if the user exists in map
	hashedUsername := Argon2Hash(username)
	userdata, ok = userDataMap[hashedUsername]
	if !ok {
		err := errors.New("[GetUser]: User not present in Datastore.")
		return nil, err
	}

	//check if the password is correct
	authPass := []byte(Argon2PasswordHash(password))
	if userlib.Equal(userdata.SymmetricKey, authPass) != true {
		err := errors.New("[GetUser]: User's password doesn't match.")
		return nil, err
	}

	//calculate newHMAC of fetched User
	macInit := userlib.NewHMAC(userdata.SymmetricKey)

	macInit.Write([]byte(userdata.Username))
	var bytes []byte
	bytes, err = json.Marshal(userdata.PrivateKey)
	if err != nil {
		return nil, err
	}
	macInit.Write(bytes)
	bytes, err = json.Marshal(userdata.FileKeys)
	if err != nil {
		return nil, err
	}
	//check if HMAC is same(not tampered)
	if userlib.Equal(macInit.Sum(nil), userdata.HMAC) != true {
		err := errors.New("[GetUser]: User's data has been tampered.")
		return nil, err
	}

	return &userdata, nil
}

// This stores a file in the datastore.
// The name of the file should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	/* TO-DO: Store encrypted filenames in FilenameMap (is it even useful?)
	* Perform Full Encryption
	 */

	// The file's MetaData is indexed into datastore by the string
	// hash(metaDataString + username + randUUID + filename)
	// The MetaData stores information about the blocks of file

	metadataIndex := metaDataString + userdata.Username + filename
	metadataIndexHashed := Argon2Hash(metadataIndex)
	_, ok := userlib.DatastoreGet(metadataIndexHashed)

	// For first store, file must not be present
	if ok {
		errString := "[StoreFile] [Argon2Key MetadataHash Collision]: " + metadataIndex + " Collided"
		panic(errString)
		return
	}

	// Random UUID in string form
	randUUID := uuid.New().String()
	fileKey := Argon2Hash(randUUID)

	// Before anything else, update the User struct with new fileKey
	userdata.FileKeys[filename] = FileSharingKey(fileKey)

	hashedUsername := Argon2Hash(userdata.Username)

	// Populate the file metadata
	var metadata MetaData
	metadata.Owner = userdata.Username
	metadata.LastEditBy = hashedUsername
	metadata.FilenameMap = make(map[string][]byte)
	metadata.FilenameMap[hashedUsername] = []byte(Argon2Hash(filename))
	metadata.GenesisUUIDNonce = uuid.New()

	genesisBlockNumber := 0
	blockIndex := fileBlocksString + userdata.Username + metadata.GenesisUUIDNonce.String() + string(genesisBlockNumber) + filename
	blockIndexHashed := Argon2Hash(blockIndex)
	metadata.GenesisBlock = blockIndexHashed

	metadata.LastUUIDNonce = metadata.GenesisUUIDNonce
	metadata.LastBlock = metadata.GenesisBlock

	// Get the HMAC of the current metadata structure
	hmac, err := MetadataHMAC(metadata, userdata.SymmetricKey)
	if err != nil {
		panic(err)
		return
	}
	metadata.HMAC = hmac

	// Marshal Metadata and store in Datastore
	bytes, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
		return
	}
	userlib.DatastoreSet(metadataIndexHashed, bytes)

	_, ok = userlib.DatastoreGet(metadata.GenesisBlock)

	// For first store, block must not be present
	if ok {
		errString := "[StoreFile] [Argon2Key BlockHash Collision]: " + blockIndex + " Collided"
		panic(errString)
		return
	}

	// TO-DO Encrypt the below struct
	var block Block
	block.Owner = metadata.Owner
	block.Content = data
	block.PrevBlockHash = ""

	// Get the HMAC of the current block structure
	hmac, err = BlockHMAC(block, userdata.SymmetricKey)
	if err != nil {
		panic(err)
		return
	}
	block.HMAC = hmac

	bytes, err = json.Marshal(block)
	if err != nil {
		panic(err)
		return
	}
	userlib.DatastoreSet(metadata.GenesisBlock, bytes)
	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	metadataIndex := metaDataString + userdata.Username + filename
	metadataIndexHashed := Argon2Hash(metadataIndex)
	val, ok := userlib.DatastoreGet(metadataIndexHashed)

	// For first store, file must not be present
	if !ok {
		errString := "Error 404 : " + filename + "not found"
		panic(errString)
		return
	}
	//Decrypt everything you encrypt
	//get file key
	//fileKey := userdata.FileKeys[filename]

	//decrypt the file-TODO
	var metadata MetaData
	json.Unmarshal(val, &metadata)

	//check the HMAC
	calcMetadataHMAC, err := MetadataHMAC(metadata, userdata.SymmetricKey)

	if !userlib.Equal(calcMetadataHMAC, metadata.HMAC) {
		errString := "Something Wrong with MetaDataHMAC"
		panic(errString)
		return
	}

	//get the file blocks and TODO- decrypt them
	val, ok = userlib.DatastoreGet(metadata.GenesisBlock)
	// check if block is present or not
	if !ok {
		errString := "Error 404 :  " + "Not Found"
		panic(errString)
		return
	}

	var block Block
	json.Unmarshal(val, &block)
	calcBlockHMAC, err := BlockHMAC(block, userdata.SymmetricKey)

	//check block  HMAC
	if !userlib.Equal(calcBlockHMAC, block.HMAC) {
		errString := "Something Wrong with BlockHMAC"
		panic(errString)
	}

	//TODO	//traverse all block until LastBlock and check their integrity
	return
}

// You may want to define what you actually want to pass as a
// sharingRecord to serialized/deserialize in the data store.
type sharingRecord struct {
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.

func (userdata *User) ShareFile(filename string, recipient string) (
	msgid string, err error) {
	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	msgid string) error {
	return nil
}

// Removes access for all others.
func (userdata *User) RevokeFile(filename string) (err error) {
	return
}
