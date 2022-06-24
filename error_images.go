package imageupsizer

import (
	"context"
	"crypto/sha512"
	"fmt"

	"github.com/kmulvey/concurrenthash"
)

var ch *concurrenthash.ConcurrentHash

var hashes = map[string]struct{}{
	"e663f9122d24f60aade166046334e60b1e195ad95a8946227e8c03cfd14031684a2f7acdcfa7322f96650259f79791c661e9b7e006735958f019c081c43bc128": {},
	"306961ff9f3c040d28bea9dfde979561efc3296999b17648fede6c7dcf9f92f0c1c79d300eb5a65a861590d6329382cf45d1666574aba2b63047fa8db14f99c8": {},
	"36d80b0d60370fcf101a95b6e4984558eed5cf84df4e0eba75a7546022556ab65ea7cf6fcceb0cdf27f0d074654e34ec693fd363b8e91eab589f0dc36961bcf4": {},
	"67cc3285a9599313c9de8773d1d51beaafab83a417cb383e20516308c6518122b8aa6483802d9445833c982fb5b1959a4442be95676bb8ce41660c4e6aac7820": {},
	"8bd63f441973551d6d5716f28cc47c7e721d1b7db26fe5c2ecabd4edef32b1ea215ff102fabd8f22dac2edadddf95c759295ebe46ec4116567ebb621c114ecaf": {},
	"9a676acc978810ea5146dda64695d06a1b83bf7c43b7a7b923c05771235dfce0b36027f6588a80dbcc9009f115c5d21799104344ce6546792b498ab4566dd315": {},
	"235f79c78d054d3de1182dab7cab9658ef61901558a4a99e7f41f869338a026fbc116d17f521f792e42d0887f4e42fe976922e40e05d841397ba393f830f7fd0": {},
	"dc987a5cee8e3cefbae339ca82cd006fefbfa9d89b82c83adb5413e1ffc91eb62c8a4fb258e7f9dbc0d4338c033aa9b205e00c0563e418e0a0d74e7ae8be6817": {},
}

func init() {
	ch = concurrenthash.NewConcurrentHash(context.Background(), 2, 2, sha512.New)
}

func isErrorImage(img *ImageData) (bool, error) {

	var hash, err = ch.HashFile(img.LocalPath)
	if err != nil {
		return false, fmt.Errorf("error hasing file: %w", err)
	}

	_, exists := hashes[hash]
	return exists, nil
}