package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	pb "khgallery/khgallery"
)

var (
	listenAddr = flag.String("listenAddr", "::", "gRPC listen address")
	listenPort = flag.Uint("listenPort", 3900, "gRPC listen port")
)

type galleryManager struct {
}

func (*galleryManager) PutPhotos(stream pb.GalleryManager_PutPhotosServer) error {

}

func getGalleryIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<")
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listenAddr, *listenPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}
