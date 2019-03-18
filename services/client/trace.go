package client

import (
	"context"
	"encoding/json"
	"reflect"

	"go-connector/services/decryption"
)

// TraceClient defines all the possible interactions with Trace.
type TraceClient interface {
	// CallTraceGql makes a call to the Trace graphql endpoint.
	CallTraceGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error
}

func (c *client) CallTraceGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error {
	err := c.callGqlEndpoint(ctx, c.urlTrace+"/graphql", query, variables, rsp)
	if err != nil {
		return err
	}

	c.decryptLinks(ctx, reflect.ValueOf(rsp))
	return nil
}

type encryptedLink struct {
	Data []byte
	Meta struct {
		Recipients []*decryption.Recipient
	}
}

// Recursively find and decrypt all links in v.
func (c *client) decryptLinks(ctx context.Context, v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = reflect.Indirect(v)
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	// If v is a slice, iterate over all of its elements.
	if v.Kind() == reflect.Slice && v.Type() != reflect.TypeOf([]byte{}) {
		for i := 0; i < v.Len(); i++ {
			c.decryptLinks(ctx, v.Index(i))
		}
		return
	}

	// If v is not a nested structure, it contains no link.
	if v.Kind() != reflect.Struct && v.Kind() != reflect.Map {
		return
	}

	c.parseAndDecryptLink(ctx, v)

	// If v is a struct, iterate over all of its fields.
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			c.decryptLinks(ctx, v.Field(i))
		}
	}

	// If v is a map, iterate over all of its values.
	if v.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			c.decryptLinks(ctx, iter.Value())
		}
	}

	return
}

// Check if the value contains a link and decrypt its data if possible.
func (c *client) parseAndDecryptLink(ctx context.Context, v reflect.Value) {

	l := v.Interface()

	// Try to parse the object into an encrypted link struct.
	// To do that, we remarshal to and unmarshal from JSON.
	var link encryptedLink
	lb, err := json.Marshal(l)
	if err != nil {
		// This is not a decryptable link.
		return
	}
	err = json.Unmarshal(lb, &link)
	if err != nil || link.Data == nil || len(link.Meta.Recipients) == 0 {
		// This is not a decryptable link.
		return
	}

	d, err := c.decryptor.DecryptLinkData(ctx, link.Data, link.Meta.Recipients)
	if err != nil {
		panic(err)
	}

	// Set the link data.
	if v.Kind() == reflect.Map {
		v.SetMapIndex(reflect.ValueOf("data"), reflect.ValueOf(d))
		return
	}

	df := v.FieldByName("Data")
	if df.Kind() == reflect.String {
		df.SetString(string(d))
	} else {
		df.Set(reflect.ValueOf(d))
	}
}
