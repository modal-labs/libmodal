package modal

import (
	"sync"

	pb "github.com/modal-labs/libmodal/modal-go/proto/modal_proto"
	"google.golang.org/grpc"
)

// SetClientFactoryForTesting overrides the gRPC client factory for tests.
// It resets global client state and returns a restore function to undo changes.
func SetClientFactoryForTesting(testClientFactory func(Profile) (grpc.ClientConnInterface, pb.ModalClientClient, error)) (restore func()) {
	origClientFactory := clientFactory
	clientFactory = testClientFactory

	// Recreate client using the overridden clientFactory, and reset the rest
	_, client, _ = clientFactory(clientProfile)
	inputPlaneClients = make(map[string]pb.ModalClientClient)
	authToken = ""

	var once sync.Once
	return func() {
		once.Do(func() {
			clientFactory = origClientFactory
			_, client, _ = clientFactory(clientProfile)
			inputPlaneClients = map[string]pb.ModalClientClient{}
			authToken = ""
		})
	}
}
