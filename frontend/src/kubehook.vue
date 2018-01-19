<template>
  <div id="kubehook">
    <b-alert v-if="error" show dismissible variant="danger" v-on:dismissed="reset">
      <p>Could not generate token: {{error}}</p>
    </b-alert>
    <b-container fluid>
      <b-row><br /></b-row>
      <b-row>
        <b-col></b-col>
        <b-col md="10">
          <b-jumbotron>
            <template slot="header">Kubernetes</template>
            <template v-if="token" slot="lead">Configure <code class="bash">kubectl</code> to use your authentication token</template>
            <template v-else slot="lead">Request a new <code class="bash">kubectl</code> authentication token</template>
            <hr class="my-4">
            <div v-if="token">
              <b-row>
                <b-col>
                  <h3>Your new authentication token</h3>
                  <pre v-highlightjs="token"><code class="bash"></code></pre>
                  <br />
                  <h3>Using your token</h3>
                  <div v-if="kubecfg">
                  <b-form inline>
                    <p>
                      Save <a :href="kubeCfgLink()">this file</a> as <code>~/.kube/config</code>
                      to automatically setup common clusters with your new
                      token.
                  </p>
                  </b-form>
                  <pre v-highlightjs="snippetKubecfg()"><code class="bash"></code></pre>
                  <br />
                  <h3>Authenticating manually</h3>
                  </div>
                  <b-form inline>
                    <p v-if="kubecfg">
                      Prefer to keep your existing <code>~/.kube/config</code>?
                      Run the following to use your token with
                      <b-input v-model="clusterID" size="sm" required />
                    </p>
                    <p v-else>
                      Run the following to use your token with
                      <b-input v-model="clusterID" size="sm" required />
                    </p>
                  </b-form>
                  <pre v-highlightjs="snippetManual()"><code class="bash"></code></pre>
                  <p>Refreshing your existing token? Run this.</p>
                  <pre v-highlightjs="snippetUpdate()"><code class="bash"></code></pre>
                </b-col>
              </b-row>
            </div>
            <div v-else>
              <b-row>
                <b-col md="3" order="1" order-md="12">
                  <b-button v-if="error" disabled block size="lg" variant="primary" v-on:click="fetchToken">Request a token</b-button>
                  <b-button v-else block size="lg" variant="primary" v-on:click="fetchToken">Request a token</b-button>
                  <br />
                </b-col>
                <b-col md="9" order="12" order-md="1">
                  <strong>Token lifetime</strong>
                  <v-slider
                    formatter="{value} days"
                    min="1"
                    max="7"
                    tooltip-dir="bottom"
                    v-model="lifetime"
                  ></v-slider>
                  <br />
                </b-col>
              </b-row>
            </div>
          </b-jumbotron>
        </b-col>
        <b-col></b-col>
      </b-row>
    </b-container>
  </div>
</template>

<script>
export default {
  name: "kubehook",
  metaInfo: {
    title: "Kubernetes Authentication",
    htmlAttrs: {
      lang: "en"
    }
  },
  data: function() {
    return {
      kubecfg: false,
      lifetime: 2,
      clusterID: "radcluster",
      token: null,
      error: null
    };
  },
  created: function() {
    this.detectKubeCfg();
  },
  methods: {
    inHours: function(lifetime) {
      return lifetime * 24 + "h";
    },
    detectKubeCfg: function() {
      var _this = this;
      this.axios
        .get("/kubecfg?lifetime=1s")
        .then(function(response) {
          _this.kubecfg = true;
        })
        .catch(function(e) {
          _this.kubecfg = false;
        });
    },
    kubeCfgLink: function() {
      return "/kubecfg?lifetime=" + this.inHours(this.lifetime);
    },
    fetchToken: function() {
      var _this = this;
      this.axios
        .post("/generate", { lifetime: this.inHours(this.lifetime) })
        .then(function(response) {
          _this.token = response.data.token;
        })
        .catch(function(e) {
          if (e.response && e.response.data.error) {
            _this.error = e.response.data.error;
            return;
          }
          if (e.request) {
            _this.error = "could not connect to API";
            return;
          }
          _this.error = e;
        });
    },
    reset: function() {
      this.error = null;
    },
    snippetKubecfg: function() {
      return (
        "# ~/.kube/config defines 'contexts' that associate your token with a cluster.\n" +
        "kubectl config get-contexts\n" +
        "\n" +
        "# Use a context to discover available namespaces.\n" +
        "kubectl --context=example get namespaces"
      );
    },
    snippetManual: function() {
      return (
        "export CLUSTER=" +
        this.clusterID +
        "\n" +
        'export TOKEN="' +
        this.token +
        '"\n' +
        "\n" +
        "# Create or update a user.\n" +
        'kubectl config set-credentials kubehook --token="${TOKEN}"\n' +
        "\n" +
        "# Associate your user with an existing cluster.\n" +
        "kubectl config set-context ${CLUSTER} --cluster=${CLUSTER} --user=kubehook\n" +
        "\n" +
        "# Use your context to discover available namespaces.\n" +
        "kubectl --context=${CLUSTER} get namespaces"
      );
    },
    snippetUpdate: function() {
      return (
        'kubectl config set-credentials kubehook --token="' + this.token + '"\n'
      );
    }
  }
};
</script>
