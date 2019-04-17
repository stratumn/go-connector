package client

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	chainscript "github.com/stratumn/go-chainscript"

	"github.com/stratumn/go-connector/services/decryption"
)

// TraceClient defines all the possible interactions with Trace.
type TraceClient interface {
	// CallTraceGql makes a call to the Trace graphql endpoint.
	CallTraceGql(ctx context.Context, query string, variables map[string]interface{}, rsp interface{}) error
	CreateLink(ctx context.Context, link *chainscript.Link) (*CreateLinkPayload, error)
	CreateLinks(ctx context.Context, links []*chainscript.Link) (*CreateLinksPayload, error)

	GetRecipientsPublicKeys(ctx context.Context, workflowID string) ([]*PublicKeyInfo, error)
	SignLink(link *chainscript.Link) error
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
	Raw  *chainscript.Link
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
		return
	}

	// If v is a map, iterate over all of its values.
	if v.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			c.decryptLinks(ctx, iter.Value())
		}
		return
	}
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
	if err != nil {
		// This is not a decryptable link.
		return
	}

	if link.Raw != nil {

		err := c.decryptor.DecryptLink(ctx, link.Raw)
		if err != nil {
			// This is not a decryptable link.
			return
		}

		// Set the decrypted raw link data.
		if v.Kind() == reflect.Map {
			v.SetMapIndex(reflect.ValueOf("raw"), reflect.ValueOf(link.Raw))
		} else {
			v.FieldByName("Raw").Set(reflect.ValueOf(link.Raw))
		}

	}

	if link.Data != nil && len(link.Meta.Recipients) != 0 {

		d, err := c.decryptor.DecryptLinkData(ctx, link.Data, link.Meta.Recipients)
		if err != nil {
			// This is not a decryptable link.
			return
		}

		// Set the link data.
		if v.Kind() == reflect.Map {
			v.SetMapIndex(reflect.ValueOf("data"), reflect.ValueOf(d))
		} else {
			df := v.FieldByName("Data")
			if df.Kind() == reflect.String {
				df.SetString(string(d))
			} else {
				df.Set(reflect.ValueOf(d))
			}
		}
	}
}

// CreateLinkPayload is the type returned by CreateLink.
type CreateLinkPayload struct {
	CreateLink struct {
		Trace struct {
			RowID string
		}
	}
}

// CreateLinkMutation is the mutation sent to create a link.
const CreateLinkMutation = `mutation CreateLinkMutation ($link: JSON!) {
	createLink(input: {link: $link}) {
		trace {
			rowId
		}
	}
}`

// CreateLink creates an attestation.
// It signs the link before sending it.
func (c *client) CreateLink(ctx context.Context, link *chainscript.Link) (*CreateLinkPayload, error) {
	err := c.SignLink(link)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{"link": link}

	var rsp CreateLinkPayload
	if err := c.CallTraceGql(ctx, CreateLinkMutation, variables, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CreateLinksPayload is the type returned by CreateLinks.
type CreateLinksPayload struct {
	CreateLinks struct {
		Links []struct {
			TraceID string
		}
	}
}

// CreateLinksMutation is the mutation sent to create multiple links.
const CreateLinksMutation = `mutation CreateLinksMutation ($links: [CreateLinkInput!]) {
	createLinks(input: $links) {
		links {
			traceId
		}
	}
}`

// CreateLinks creates multiple attestations.
// It signs the links before sending them.
func (c *client) CreateLinks(ctx context.Context, links []*chainscript.Link) (*CreateLinksPayload, error) {
	linksInput := make([]map[string]interface{}, len(links))
	for i, link := range links {
		err := c.SignLink(link)
		if err != nil {
			return nil, err
		}
		linksInput[i] = map[string]interface{}{"link": link}
	}

	variables := map[string]interface{}{"links": linksInput}

	var rsp CreateLinksPayload
	if err := c.CallTraceGql(ctx, CreateLinksMutation, variables, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// SignLink signs a link.
func (c *client) SignLink(link *chainscript.Link) error {
	for _, sig := range link.Signatures {
		if bytes.Equal(sig.PublicKey, c.signingPublicKey) {
			return nil
		}
	}
	return link.Sign(c.signingPrivateKey, "[version,data,meta]")
}

// PublicKeyInfo contains the public key and its ID.
type PublicKeyInfo struct {
	ID        string
	PublicKey []byte
}

// RecipientsKeysQuery is the query sent to fetch the public
// keys of the workflow participants from trace API.
const RecipientsKeysQuery = `query GetRecipientsKeysQuery($workflowId: BigInt!) {
	workflowByRowId(rowId: $workflowId) {
		groups {
			nodes {
				owner {
					encryptionKey { rowId, publicKey }
				}
			}
		}
	}
}`

// RecipientsKeysRsp is the structure of the response of the
// recipientsKeysQuery query.
type RecipientsKeysRsp struct {
	WorkflowByRowID *struct {
		Groups struct {
			Nodes []struct {
				Owner struct {
					EncryptionKey struct {
						PublicKey string
						RowID     string
					}
				}
			}
		}
	}
}

// GetRecipientsPublicKeys gets the public keys of the workflow's group owners.
// the keys are cached for `keyRefreshInterval` minutes.
func (c *client) GetRecipientsPublicKeys(ctx context.Context, workflowID string) ([]*PublicKeyInfo, error) {

	variables := map[string]interface{}{"workflowId": workflowID}

	rsp := RecipientsKeysRsp{}
	if err := c.CallTraceGql(ctx, RecipientsKeysQuery, variables, &rsp); err != nil {
		return nil, err
	}
	if rsp.WorkflowByRowID == nil {
		return nil, errors.Errorf("workflow %s not found", workflowID)
	}

	res := make([]*PublicKeyInfo, len(rsp.WorkflowByRowID.Groups.Nodes))
	for i, g := range rsp.WorkflowByRowID.Groups.Nodes {
		pk := []byte(g.Owner.EncryptionKey.PublicKey)
		res[i] = &PublicKeyInfo{PublicKey: pk, ID: g.Owner.EncryptionKey.RowID}
	}

	return res, nil
}
