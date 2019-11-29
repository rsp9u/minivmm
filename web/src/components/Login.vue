<template lang="pug">
  #login_card
    v-app
      v-content
        v-container
          v-card.mx-auto(max-width="300").login-card
            v-card-text
              h2.login-title minivmm
              v-btn(v-on:click="login") OIDC Login
</template>

<script>
import axios from "axios";

import Oidc from "oidc-client";
import c from "@/const";
import util from "@/util";

const redirectURI = util.locationOrigin() + "/api/v1/login";

export default {
  name: "Login",
  data() {
    return {
      user: null,
      signedIn: false
    };
  },
  methods: {
    login() {
      axios
        .get(util.locationOrigin() + "/api/v1/auth")
        .then(response => this.oidc_auth(response.data.oidc_url))
        .catch(err => this.oidc_auth(err.response.data.oidc_url));
    },
    oidc_auth(oidc_provider_url) {
      if (oidc_provider_url === undefined || oidc_provider_url === "") {
        alert("oidc provider url is not set.");
        return;
      }
      const userManager = new Oidc.UserManager({
        userStore: new Oidc.WebStorageStateStore(),
        authority: oidc_provider_url,
        client_id: c.OIDC_CLIENT_ID,
        redirect_uri: redirectURI,
        response_type: "code",
        scope: "openid",
        metadata: {
          issuer: oidc_provider_url + "/",
          authorization_endpoint: oidc_provider_url + "/oauth2/auth",
          userinfo_endpoint: oidc_provider_url + "/userinfo",
          end_session_endpoint: oidc_provider_url + "/oauth2/sessions/logout",
          jwks_uri: oidc_provider_url + "/.well-known/jwks.json"
        }
      });
      userManager.signinRedirect().catch(err => console.log(err));
    }
  }
};
</script>

<style lang="sass">
.login-title
  align: center
  padding: 10px

.login-card
  align: center
  width: 20em
</style>
