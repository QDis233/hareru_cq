package hareru_cq

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func getHttpRes(url string) ([]byte, error) {
	client := http.Client{}
	response, err := client.Get(url)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http request error: %d %s", response.StatusCode, response.Status))
	}

	return io.ReadAll(response.Body)
}
