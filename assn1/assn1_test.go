package assn1

import "github.com/fenilfadadu/CS628-assn1/userlib"
import "testing"
import "reflect"
import (
	"crypto/sha256"
	"encoding/hex"
)

// You can actually import other stuff if you want IN YOUR TEST
// HARNESS ONLY.  Note that this is NOT considered part of your
// solution, but is how you make sure your solution is correct.

func TestInit(t *testing.T) {
	t.Logf("%x \n", sha256.Sum256([]byte("app")))
	temp := hex.EncodeToString(userlib.Argon2Key([]byte("app"), nil, 32))
	t.Logf("%v \n", temp)
	t.Logf("%x \n", userlib.Argon2Key([]byte("app"), nil, 32))
	t.Logf("%x \n", userlib.Argon2Key([]byte("app"), nil, 32))
	t.Log("Initialization test")
	userlib.DebugPrint = true
	//	someUsefulThings()

	userlib.DebugPrint = false
	aliceUser := "alice"
	alicePass := "foobar"
	u, err := InitUser(aliceUser, alicePass)
	if err != nil {
		t.Error("Got InitUser Error", err)
	} else {
		t.Logf("Username =  %s\n", u.Username)
		t.Logf("HMAC = %x", u.HMAC)
	}
	// key is private key
	key, err := userlib.GenerateRSAKey()
	if err != nil {
		t.Error("Got RSA error", err)
	}
	userlib.KeystoreSet(aliceUser, key.PublicKey)

	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
	}
	// t.Log() only produces output if you run with "go test -v"
	// You probably want many more tests here.
}

func TestStorage(t *testing.T) {
	// And some more tests, because
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	t.Log("Loaded user", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
	}
}

func TestShare(t *testing.T) {
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
	}

	var v, v2 []byte
	var msgid string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
	}

	msgid, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
	}
	err = u2.ReceiveFile("file2", "alice", msgid)
	if err != nil {
		t.Error("Failed to receive the share message", err)
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
	}

}
