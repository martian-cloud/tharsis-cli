/*

Portions of this file are from
https://github.com/hashicorp/terraform/blob/6b290cf163f6816686bf0a5ae85ff4cb37b3beed/internal/command/login.go.

Copyright (c) 2022 HashiCorp. All rights reserved.

Mozilla Public License, version 2.0

1. Definitions

1.1. “Contributor”

	means each individual or legal entity that creates, contributes to the
	creation of, or owns Covered Software.

1.2. “Contributor Version”

	means the combination of the Contributions of others (if any) used by a
	Contributor and that particular Contributor’s Contribution.

1.3. “Contribution”

	means Covered Software of a particular Contributor.

1.4. “Covered Software”

	means Source Code Form to which the initial Contributor has attached the
	notice in Exhibit A, the Executable Form of such Source Code Form, and
	Modifications of such Source Code Form, in each case including portions
	thereof.

1.5. “Incompatible With Secondary Licenses”

	means

	a. that the initial Contributor has attached the notice described in
	   Exhibit B to the Covered Software; or

	b. that the Covered Software was made available under the terms of version
	   1.1 or earlier of the License, but not also under the terms of a
	   Secondary License.

1.6. “Executable Form”

	means any form of the work other than Source Code Form.

1.7. “Larger Work”

	means a work that combines Covered Software with other material, in a separate
	file or files, that is not Covered Software.

1.8. “License”

	means this document.

1.9. “Licensable”

	means having the right to grant, to the maximum extent possible, whether at the
	time of the initial grant or subsequently, any and all of the rights conveyed by
	this License.

1.10. “Modifications”

	means any of the following:

	a. any file in Source Code Form that results from an addition to, deletion
	   from, or modification of the contents of Covered Software; or

	b. any new file in Source Code Form that contains any Covered Software.

1.11. “Patent Claims” of a Contributor

	means any patent claim(s), including without limitation, method, process,
	and apparatus claims, in any patent Licensable by such Contributor that
	would be infringed, but for the grant of the License, by the making,
	using, selling, offering for sale, having made, import, or transfer of
	either its Contributions or its Contributor Version.

1.12. “Secondary License”

	means either the GNU General Public License, Version 2.0, the GNU Lesser
	General Public License, Version 2.1, the GNU Affero General Public
	License, Version 3.0, or any later versions of those licenses.

1.13. “Source Code Form”

	means the form of the work preferred for making modifications.

1.14. “You” (or “Your”)

	means an individual or a legal entity exercising rights under this
	License. For legal entities, “You” includes any entity that controls, is
	controlled by, or is under common control with You. For purposes of this
	definition, “control” means (a) the power, direct or indirect, to cause
	the direction or management of such entity, whether by contract or
	otherwise, or (b) ownership of more than fifty percent (50%) of the
	outstanding shares or beneficial ownership of such entity.

2. License Grants and Conditions

2.1. Grants

	Each Contributor hereby grants You a world-wide, royalty-free,
	non-exclusive license:

	a. under intellectual property rights (other than patent or trademark)
	   Licensable by such Contributor to use, reproduce, make available,
	   modify, display, perform, distribute, and otherwise exploit its
	   Contributions, either on an unmodified basis, with Modifications, or as
	   part of a Larger Work; and

	b. under Patent Claims of such Contributor to make, use, sell, offer for
	   sale, have made, import, and otherwise transfer either its Contributions
	   or its Contributor Version.

2.2. Effective Date

	The licenses granted in Section 2.1 with respect to any Contribution become
	effective for each Contribution on the date the Contributor first distributes
	such Contribution.

2.3. Limitations on Grant Scope

	The licenses granted in this Section 2 are the only rights granted under this
	License. No additional rights or licenses will be implied from the distribution
	or licensing of Covered Software under this License. Notwithstanding Section
	2.1(b) above, no patent license is granted by a Contributor:

	a. for any code that a Contributor has removed from Covered Software; or

	b. for infringements caused by: (i) Your and any other third party’s
	   modifications of Covered Software, or (ii) the combination of its
	   Contributions with other software (except as part of its Contributor
	   Version); or

	c. under Patent Claims infringed by Covered Software in the absence of its
	   Contributions.

	This License does not grant any rights in the trademarks, service marks, or
	logos of any Contributor (except as may be necessary to comply with the
	notice requirements in Section 3.4).

2.4. Subsequent Licenses

	No Contributor makes additional grants as a result of Your choice to
	distribute the Covered Software under a subsequent version of this License
	(see Section 10.2) or under the terms of a Secondary License (if permitted
	under the terms of Section 3.3).

2.5. Representation

	Each Contributor represents that the Contributor believes its Contributions
	are its original creation(s) or it has sufficient rights to grant the
	rights to its Contributions conveyed by this License.

2.6. Fair Use

	This License is not intended to limit any rights You have under applicable
	copyright doctrines of fair use, fair dealing, or other equivalents.

2.7. Conditions

	Sections 3.1, 3.2, 3.3, and 3.4 are conditions of the licenses granted in
	Section 2.1.

3. Responsibilities

3.1. Distribution of Source Form

	All distribution of Covered Software in Source Code Form, including any
	Modifications that You create or to which You contribute, must be under the
	terms of this License. You must inform recipients that the Source Code Form
	of the Covered Software is governed by the terms of this License, and how
	they can obtain a copy of this License. You may not attempt to alter or
	restrict the recipients’ rights in the Source Code Form.

3.2. Distribution of Executable Form

	If You distribute Covered Software in Executable Form then:

	a. such Covered Software must also be made available in Source Code Form,
	   as described in Section 3.1, and You must inform recipients of the
	   Executable Form how they can obtain a copy of such Source Code Form by
	   reasonable means in a timely manner, at a charge no more than the cost
	   of distribution to the recipient; and

	b. You may distribute such Executable Form under the terms of this License,
	   or sublicense it under different terms, provided that the license for
	   the Executable Form does not attempt to limit or alter the recipients’
	   rights in the Source Code Form under this License.

3.3. Distribution of a Larger Work

	You may create and distribute a Larger Work under terms of Your choice,
	provided that You also comply with the requirements of this License for the
	Covered Software. If the Larger Work is a combination of Covered Software
	with a work governed by one or more Secondary Licenses, and the Covered
	Software is not Incompatible With Secondary Licenses, this License permits
	You to additionally distribute such Covered Software under the terms of
	such Secondary License(s), so that the recipient of the Larger Work may, at
	their option, further distribute the Covered Software under the terms of
	either this License or such Secondary License(s).

3.4. Notices

	You may not remove or alter the substance of any license notices (including
	copyright notices, patent notices, disclaimers of warranty, or limitations
	of liability) contained within the Source Code Form of the Covered
	Software, except that You may alter any license notices to the extent
	required to remedy known factual inaccuracies.

3.5. Application of Additional Terms

	You may choose to offer, and to charge a fee for, warranty, support,
	indemnity or liability obligations to one or more recipients of Covered
	Software. However, You may do so only on Your own behalf, and not on behalf
	of any Contributor. You must make it absolutely clear that any such
	warranty, support, indemnity, or liability obligation is offered by You
	alone, and You hereby agree to indemnify every Contributor for any
	liability incurred by such Contributor as a result of warranty, support,
	indemnity or liability terms You offer. You may include additional
	disclaimers of warranty and limitations of liability specific to any
	jurisdiction.

4. Inability to Comply Due to Statute or Regulation

	If it is impossible for You to comply with any of the terms of this License
	with respect to some or all of the Covered Software due to statute, judicial
	order, or regulation then You must: (a) comply with the terms of this License
	to the maximum extent possible; and (b) describe the limitations and the code
	they affect. Such description must be placed in a text file included with all
	distributions of the Covered Software under this License. Except to the
	extent prohibited by statute or regulation, such description must be
	sufficiently detailed for a recipient of ordinary skill to be able to
	understand it.

5. Termination

5.1. The rights granted under this License will terminate automatically if You

	fail to comply with any of its terms. However, if You become compliant,
	then the rights granted under this License from a particular Contributor
	are reinstated (a) provisionally, unless and until such Contributor
	explicitly and finally terminates Your grants, and (b) on an ongoing basis,
	if such Contributor fails to notify You of the non-compliance by some
	reasonable means prior to 60 days after You have come back into compliance.
	Moreover, Your grants from a particular Contributor are reinstated on an
	ongoing basis if such Contributor notifies You of the non-compliance by
	some reasonable means, this is the first time You have received notice of
	non-compliance with this License from such Contributor, and You become
	compliant prior to 30 days after Your receipt of the notice.

5.2. If You initiate litigation against any entity by asserting a patent

	infringement claim (excluding declaratory judgment actions, counter-claims,
	and cross-claims) alleging that a Contributor Version directly or
	indirectly infringes any patent, then the rights granted to You by any and
	all Contributors for the Covered Software under Section 2.1 of this License
	shall terminate.

5.3. In the event of termination under Sections 5.1 or 5.2 above, all end user

	license agreements (excluding distributors and resellers) which have been
	validly granted by You or Your distributors under this License prior to
	termination shall survive termination.

6. Disclaimer of Warranty

	Covered Software is provided under this License on an “as is” basis, without
	warranty of any kind, either expressed, implied, or statutory, including,
	without limitation, warranties that the Covered Software is free of defects,
	merchantable, fit for a particular purpose or non-infringing. The entire
	risk as to the quality and performance of the Covered Software is with You.
	Should any Covered Software prove defective in any respect, You (not any
	Contributor) assume the cost of any necessary servicing, repair, or
	correction. This disclaimer of warranty constitutes an essential part of this
	License. No use of  any Covered Software is authorized under this License
	except under this disclaimer.

7. Limitation of Liability

	Under no circumstances and under no legal theory, whether tort (including
	negligence), contract, or otherwise, shall any Contributor, or anyone who
	distributes Covered Software as permitted above, be liable to You for any
	direct, indirect, special, incidental, or consequential damages of any
	character including, without limitation, damages for lost profits, loss of
	goodwill, work stoppage, computer failure or malfunction, or any and all
	other commercial damages or losses, even if such party shall have been
	informed of the possibility of such damages. This limitation of liability
	shall not apply to liability for death or personal injury resulting from such
	party’s negligence to the extent applicable law prohibits such limitation.
	Some jurisdictions do not allow the exclusion or limitation of incidental or
	consequential damages, so this exclusion and limitation may not apply to You.

8. Litigation

	Any litigation relating to this License may be brought only in the courts of
	a jurisdiction where the defendant maintains its principal place of business
	and such litigation shall be governed by laws of that jurisdiction, without
	reference to its conflict-of-law provisions. Nothing in this Section shall
	prevent a party’s ability to bring cross-claims or counter-claims.

9. Miscellaneous

	This License represents the complete agreement concerning the subject matter
	hereof. If any provision of this License is held to be unenforceable, such
	provision shall be reformed only to the extent necessary to make it
	enforceable. Any law or regulation which provides that the language of a
	contract shall be construed against the drafter shall not be used to construe
	this License against a Contributor.

10. Versions of the License

10.1. New Versions

	Mozilla Foundation is the license steward. Except as provided in Section
	10.3, no one other than the license steward has the right to modify or
	publish new versions of this License. Each version will be given a
	distinguishing version number.

10.2. Effect of New Versions

	You may distribute the Covered Software under the terms of the version of
	the License under which You originally received the Covered Software, or
	under the terms of any subsequent version published by the license
	steward.

10.3. Modified Versions

	If you create software not governed by this License, and you want to
	create a new license for such software, you may create and use a modified
	version of this License if you rename the license and remove any
	references to the name of the license steward (except to note that such
	modified license differs from this License).

10.4. Distributing Source Code Form that is Incompatible With Secondary Licenses

	If You choose to distribute Source Code Form that is Incompatible With
	Secondary Licenses under the terms of this version of the License, the
	notice described in Exhibit B of this License must be attached.

Exhibit A - Source Code Form License Notice

	This Source Code Form is subject to the
	terms of the Mozilla Public License, v.
	2.0. If a copy of the MPL was not
	distributed with this file, You can
	obtain one at
	http://mozilla.org/MPL/2.0/.

If it is not possible or desirable to put the notice in a particular file, then
You may include the notice in a location (such as a LICENSE file in a relevant
directory) where a recipient would be likely to look for such a notice.

You may add additional accurate notices of copyright ownership.

Exhibit B - “Incompatible With Secondary Licenses” Notice

	This Source Code Form is “Incompatible
	With Secondary Licenses”, as defined by
	the Mozilla Public License, v. 2.0.

Tharsis modifications:

	- Some portions of the original file from line #365 have been
	adapted to create the SSO login command for the Tharsis CLI.
*/

package command

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	uuid "github.com/hashicorp/go-uuid"
	"github.com/mitchellh/cli"
	"github.com/pkg/browser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"golang.org/x/oauth2"
)

const (

	// Port range for temporary web server.
	tempServerMinPort = 21000
	tempServerMaxPort = 21200

	// Header reading timeout for temporary web server.
	readHeaderTimeout = 30 * time.Second

	// originHeader is added to the request during the token exchange.
	originHeader = "http://localhost"

	// Path for "well-known" URL.
	wellKnownURLPath = "/.well-known/terraform.json"
)

// customHeaderTransport is used to set custom header on the
// token exchange requests with the IDP.
type customHeaderTransport struct {
	rt                  http.RoundTripper
	includeOriginHeader bool
}

// RoundTrip adds a custom 'Origin' header to the request for PKCE flow.
func (t *customHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.includeOriginHeader {
		req.Header.Set("Origin", originHeader)
	}
	return t.rt.RoundTrip(req)
}

// loginCommand is the top-level structure for the login command.
type loginCommand struct {
	meta *Metadata
}

// NewLoginCommandFactory returns a loginCommand struct.
func NewLoginCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return loginCommand{
			meta: meta,
		}, nil
	}
}

func (lc loginCommand) Run(args []string) int {
	lc.meta.Logger.Debugf("Starting the 'login' command with %d arguments:", len(args))
	for ix, arg := range args {
		lc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	_, cmdArgs, err := optparser.ParseCommandOptions(lc.meta.BinaryName+" login",
		optparser.OptionDefinitions{}, args)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to parse login options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive login arguments: %s", cmdArgs)
		lc.meta.Logger.Error(output.FormatError(msg, nil), lc.HelpLogin())
		return 1
	}

	// Cannot delay reading settings past this point.
	currentSettings, err := lc.meta.ReadSettings()
	if err != nil {
		if err != settings.ErrNoSettings || lc.meta.DefaultEndpointURL == "" {
			lc.meta.Logger.Error(output.FormatError("Failed to read settings file", err))
			return 1
		}
		// Build settings when they don't exist.
		profile := settings.Profile{TharsisURL: lc.meta.DefaultEndpointURL}
		profiles := map[string]settings.Profile{"default": profile}
		currentSettings = &settings.Settings{Profiles: profiles, CurrentProfile: &profile}
	}

	// Use the 'well-known' endpoint to get the OAuth2 configuration.
	oauthCfg, err := lc.fetchAuthConfig(currentSettings.CurrentProfile.TharsisURL)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to fetch OAuth config", err))
		return 1
	}

	// Build the request state UUID.
	requestState, err := buildRequestState()
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to build request state", err))
		return 1
	}

	// Build the proof key and challenge values.
	proofKey, proofKeyChallenge, err := buildProofKeyChallenge()
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to build proof key challenge", err))
		return 1
	}

	// Open the net.listener and build the callback URL.
	netListener, callbackURL, err := openNetListener()
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to create temporary web server", err))
		return 1
	}

	// Launch the embedded web server.
	webServerChannel, server, err := lc.launchWebServer(requestState, proofKey, netListener)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to launch temporary web server", err))
		return 1
	}

	// Create the authCodeURL from the OAuth2 configuration.
	oauthCfg.RedirectURL = callbackURL
	authCodeURL := oauthCfg.AuthCodeURL(
		requestState,
		oauth2.SetAuthURLParam("code_challenge", proofKeyChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Launch the web browser.
	err = lc.launchBrowser(authCodeURL, currentSettings)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to launch web browser for OAuth redirect", err))
		return 1
	}

	settingsFilePath, err := settings.DefaultSettingsFilename()
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to get default settings file name", err))
		return 1
	}
	lc.meta.UI.Output("Tharsis has opened your default browser to complete the SSO login.")
	lc.meta.UI.Output("")
	lc.meta.UI.Output("If the login is successful, Tharsis will store the authentication token in:")
	lc.meta.UI.Output(settingsFilePath)

	// Capture the token.  This also terminates the web server.
	token, err := lc.captureToken(oauthCfg, proofKey, webServerChannel, server)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to capture token in temporary web server", err))
		return 1
	}

	// Store the token.
	err = lc.storeToken(token, currentSettings)
	if err != nil {
		lc.meta.Logger.Error(output.FormatError("failed to store token in settings file", err))
		return 1
	}

	lc.meta.UI.Output(fmt.Sprintf("\nTharsis SSO login was successful using the %s profile.\n",
		lc.meta.CurrentProfileName))
	return 0
}

func (lc loginCommand) Synopsis() string {
	return "Log in to the OAuth2 provider and return an authentication token."
}

func (lc loginCommand) Help() string {
	return lc.HelpLogin()
}

// HelpLogin produces the long (many lines) help string for the 'login' command.
func (lc loginCommand) HelpLogin() string {
	return fmt.Sprintf(`
Usage: %s [global options] sso login [options]

   The login command starts an embedded web server and opens
   a web browser page or tab pointed at said web server.
   That redirects to the OAuth2 provider's login page, where
   the user can sign in. If there is an SSO scheme active,
   that will sign in the user. The login command captures
   the authentication token for use in subsequent commands.

`, lc.meta.BinaryName)
}

// Fetch the 'well-known' URL from the settings.profile.TharsisURL to get the auth URL.
func (lc loginCommand) fetchAuthConfig(tharsisURL string) (*oauth2.Config, error) {
	asURL, err := url.Parse(tharsisURL)
	if err != nil {
		return nil, err
	}
	wellKnownURL := url.URL{
		Scheme: asURL.Scheme,
		Host:   asURL.Host,
		Path:   wellKnownURLPath,
	}
	lc.meta.Logger.Debugf("will fetch well-known URL: %s", wellKnownURL.String())

	resp, err := http.Get(wellKnownURL.String()) // nosemgrep: gosec.G107-1
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d from well-known URL", resp.StatusCode)
	}

	var bodyStruct struct {
		LoginDotV1 struct {
			AuthZ      string   `json:"authz"`
			Client     string   `json:"client"`
			Token      string   `json:"token"`
			GrantTypes []string `json:"grant_types"`
			Scopes     []string `json:"scopes"`
		} `json:"login.v1"`
		// Ignore the rest of the fields.
		/*
			ModulesDotV1 string `json"modules.v1"`
			StateDotV2   string `json"state.v2"`
			TFEDotV2     string `json"tfe.v2"`
			TFEDotV2Dot1 string `json"tfe.v2.1"`
			TFEDotV2Dot2 string `json"tfe.v2.2"`
		*/
	}
	err = json.Unmarshal(body, &bodyStruct)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID: bodyStruct.LoginDotV1.Client,
		Endpoint: oauth2.Endpoint{
			AuthURL:  bodyStruct.LoginDotV1.AuthZ,
			TokenURL: bodyStruct.LoginDotV1.Token,
		},
		Scopes: bodyStruct.LoginDotV1.Scopes,
	}, nil
}

// buildRequestState generates the request state UUID.
func buildRequestState() (string, error) {
	result, err := uuid.GenerateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID, likely not enough pseudo-random entropy: %s", err)
	}
	return result, nil
}

// buildProofKeyChallenge generates the proof key and challenge values.
func buildProofKeyChallenge() (string, string, error) {
	firstUUID, err := uuid.GenerateUUID()
	if err != nil {
		return "", "", err
	}

	randomInt, err := rand.Int(rand.Reader, big.NewInt(999999999))
	if err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%s.%09d", firstUUID, randomInt)

	hasher := sha256.New()
	hasher.Write([]byte(key))
	challenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return key, challenge, nil
}

// openNetListener builds a net.listener and callback URL.
func openNetListener() (net.Listener, string, error) {
	for port := tempServerMinPort; port < tempServerMaxPort; port++ {
		listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			// Lack of an error means we succeeded in opening the listener on the port.
			// No trailing slash.
			//
			// The callback URL must use "localhost", not "127.0.0.1".
			callbackURL := fmt.Sprintf("http://localhost:%d/login", port)
			return listener, callbackURL, nil
		}
	}

	// Getting here means no port was available.
	return nil, "", fmt.Errorf("no port could be opened for the temporary web server")
}

// launchWebServer launches the embedded web server and returns the termination channel.
func (lc loginCommand) launchWebServer(requestState, _ string,
	netListener net.Listener) (chan string, *http.Server, error) {
	lc.meta.Logger.Debug("callback server: creating server")
	codeChannel := make(chan string)
	httpServer := &http.Server{
		ReadHeaderTimeout: readHeaderTimeout,
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			lc.meta.Logger.Debug("callback server: handler called")

			// Parse the form.
			err := req.ParseForm()
			if err != nil {
				lc.meta.Logger.Errorf("callback server: cannot parse form on callback request: %s", err)
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			// Check that the request state matches.
			gotState := req.Form.Get("state")
			lc.meta.Logger.Debugf("callback server got state: %s", gotState)

			if gotState != requestState {
				lc.meta.Logger.Debugf("request with incorrect state: %#v", req)
				lc.meta.Logger.Debugf("URL of request with incorrect state: %#v", req.URL)
				lc.meta.Logger.Debugf("header of request with incorrect state: %#v", req.Header)
				if gotState == "" {
					// Ignore spurious requests (such as for favicon.ico) without state.
					return
				}
				lc.meta.Logger.Error("callback server: incorrect 'state' value")
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			// Did the response return a code?
			gotCode := req.Form.Get("code")
			if gotCode == "" {
				lc.meta.Logger.Error("callback server: no 'code' argument in callback request")
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			// Send the code back to the main execution line.
			lc.meta.Logger.Debugf("callback server: got an authorization code: %s", gotCode)
			codeChannel <- gotCode
			close(codeChannel)
			lc.meta.Logger.Debug("callback server: sent authorization code to channel; closed channel.")

			// Return an HTTP response.
			lc.meta.Logger.Debug("callback server: returning an HTTP response")
			resp.Header().Add("Content-Type", "text/html")
			resp.WriteHeader(http.StatusOK)
			_, _ = resp.Write([]byte(lc.buildCallbackResponseBody()))
		}),
	}

	// After creating the server, launch it.
	go func() {
		// At this point, Terraform does
		// defer logging.PanicHandler()
		err := httpServer.Serve(netListener)
		if err != nil && err != http.ErrServerClosed {
			lc.meta.Logger.Error("failed to start the temporary login server")
			close(codeChannel)
		}
	}()

	return codeChannel, httpServer, nil
}

// buildCallbackResponseBody builds the response body to be returned by callback server.
func (lc loginCommand) buildCallbackResponseBody() string {
	return fmt.Sprintf(`
	<html>
	<head>
	<title>%s Login</title>
	<style type="text/css">
	body {
		font-family: monospace;
		color: #fff;
		background-color: #000;
	}
	</style>
	</head>
	<body>
	<p>The %s SSO login has successfully completed. This page can now be closed.</p>
	</body>
	</html>
	`, lc.meta.DisplayTitle, lc.meta.DisplayTitle)
}

func (lc loginCommand) launchBrowser(authCodeURL string, currentSettings *settings.Settings) error {

	// Assume it is possible to launch a browser from here.
	asURL, err := url.Parse(currentSettings.CurrentProfile.TharsisURL)
	if err != nil {
		return err
	}
	lc.meta.UI.Output(fmt.Sprintf("\n%s must now open a web browser to the login page for host %s\n",
		lc.meta.DisplayTitle, asURL.Host))

	// Cover the bases in case the browser does not fly.
	lc.meta.UI.Output(fmt.Sprintf("If a browser does not open automatically, open the following URL:\n%s\n",
		authCodeURL))

	// Ask the user to wait.
	lc.meta.UI.Output(fmt.Sprintf("%s will now wait for the host to signal that login was successful.\n\n",
		lc.meta.DisplayTitle))

	// Either it will fly or it won't.
	return browser.OpenURL(authCodeURL)
}

// captureToken captures the token and terminates the temporary web server.
func (lc loginCommand) captureToken(oauthCfg *oauth2.Config, proofKey string,
	webServerChannel chan string, server *http.Server) (*oauth2.Token, error) {

	// Wait for a code (or signal that no code is coming).
	code, ok := <-webServerChannel
	if !ok {
		// No code came or ever will come.
		return nil, fmt.Errorf("it was not possible to capture a token")
	}

	// Immediately terminate the web server.
	err := lc.terminateWebServer(server)
	if err != nil {
		return nil, err
	}

	token, err := lc.exchangeAuthCodeForToken(oauthCfg, code, proofKey)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain an authentication token: %w", err)
	}

	return token, nil
}

func (lc loginCommand) exchangeAuthCodeForToken(oauthCfg *oauth2.Config, code string, proofKey string) (*oauth2.Token, error) {
	var token *oauth2.Token
	var err error
	// Create an HTTP client to do the exchanging.
	// Some of these parameters might not be essential, but this is what Terraform uses.
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	customTransport := &customHeaderTransport{
		rt: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
		includeOriginHeader: true,
	}

	httpClient := &http.Client{
		// Use a custom transport to add 'Origin' header to each request.
		Transport: customTransport,
	}

	// Add custom http client to context
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)

	token, err = oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", proofKey))

	// Depending on how the app registration is configured, Azure IDP will return the following error code if the origin header
	// is included; therefore, we'll retry the request without the origin header
	if err != nil && strings.Contains(err.Error(), "AADSTS9002326") {
		customTransport.includeOriginHeader = false
		token, err = oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", proofKey))
	}

	return token, err
}

// storeToken stores the token.
func (lc loginCommand) storeToken(token *oauth2.Token, currentSettings *settings.Settings) error {

	// Update the internal data structure.
	if currentSettings.CurrentProfile == nil {
		return fmt.Errorf("nil current profile, cannot store token")
	}

	// Must find the name of the current profile in order to store the token.
	var foundName string
	for name, candidate := range currentSettings.Profiles {
		if candidate == *currentSettings.CurrentProfile {
			// found it
			foundName = name
		}
	}
	if foundName == "" {
		return fmt.Errorf("failed to find name of current profile")
	}
	foundProfile := currentSettings.Profiles[foundName] // returns a copy of the struct
	foundProfile.Token = &token.AccessToken             // writes to the copy of the struct
	currentSettings.Profiles[foundName] = foundProfile  // replaces the original struct

	// Write the updated file.
	err := currentSettings.WriteSettingsFile(nil)
	if err != nil {
		return fmt.Errorf("failed to write updated credentials: %s", err)
	}

	return nil
}

// terminateWebServer terminates the embedded web server.
func (lc loginCommand) terminateWebServer(server *http.Server) error {
	// Terraform calls server.Close(), but Shutdown() is recommended as being more graceful.
	err := server.Shutdown(context.Background())
	if err != nil {
		lc.meta.Logger.Info("Warning: Unable to close the temporary callback HTTP server.")
		return err
	}

	return nil
}
