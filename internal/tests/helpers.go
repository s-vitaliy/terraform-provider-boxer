package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"io"
	"net/http"
	"net/url"
	"terraform-provider-boxer/internal"
	"text/template"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetExternalToken(services *Services) (string, error) {
	form := url.Values{}
	form.Add("client_id", services.ExternalIdp.Credentials.ClientID)
	form.Add("client_secret", services.ExternalIdp.Credentials.ClientSecret)
	form.Add("username", services.ExternalIdp.Credentials.Username)
	form.Add("password", services.ExternalIdp.Credentials.Password)
	form.Add("grant_type", services.ExternalIdp.Credentials.GrantType)

	endpoint := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", services.ExternalIdp.Endpoint)
	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return "", fmt.Errorf("failed get token from external identity provider: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("failed to unmapshal response: %w", err)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("got empty token in response: %s", body)
	}
	return tr.AccessToken, nil
}

func providerFactory() tfprotov6.ProviderServer {
	p, _ := providerserver.NewProtocol6WithError(internal.BoxerProvider{})()
	return p
}

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"boxer": func() (tfprotov6.ProviderServer, error) {
		return providerFactory(), nil
	},
}

func RenderTemplate(testContext *TestContext, templateName string) string {
	tpl, err := template.New("configuration").ParseFiles(fmt.Sprintf("templates/%s", templateName))
	if err != nil {
		panic(err)
	}

	fmt.Println("Generating test configuration...")
	var buf bytes.Buffer
	err = tpl.ExecuteTemplate(&buf, templateName, testContext)
	if err != nil {
		panic(err)
	}
	result := buf.String()
	return result
}
