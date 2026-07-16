package ytmusicclient

import (
	"flag"
	"log"

	musicpb "github.com/kumneger0/clispot/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func GetYtMusicClient() (musicpb.MusicServiceClient, *grpc.ClientConn) {
	flag.Parse()
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := musicpb.NewMusicServiceClient(conn)
	return client, conn
}
