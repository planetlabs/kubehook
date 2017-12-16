<template>
  <div id="kubehook">
    <b-alert v-if="error" show dismissible variant="danger">
      <p>Could not generate token: {{error}}</p>
    </b-alert>
    <b-container fluid>
      <b-row><br /></b-row>
      <b-row>
        <b-col></b-col>
        <b-col md="10">
          <b-jumbotron>
            <template slot="header">Kubernetes</template>
            <template v-if="token" slot="lead">Configure <code class="bash">kubectl</code> to use your authentication token</pre></p></template>
            <template v-else slot="lead">Request a new <code class="bash">kubectl</code> authentication token</pre></p></template>
            <hr class="my-4">
            <div v-if="token">
              <b-row>
                <b-col>
                  <h3>Your new authentication token</h3>
                  <pre v-highlightjs="token"><code class="bash"></code></pre>
                  <br />
                  <h3>Using your token</h3>
                  <b-form inline>
                    <p>
                    Run the following commands to use your new token with
                    <b-input v-model="clusterID" size="sm" required />
                  </p>
                  </b-form>
                  <pre v-highlightjs="shellSnippet()"><code class="bash"></code></pre>
                </b-col>
              </b-row>
            </div>
            <div v-else>
              <b-row>
                <b-col md="3" order="1" order-md="12">
                  <b-button block size="lg" variant="primary" v-on:click="fetchToken">Request a token</b-button>
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
  name: 'kubehook',
  metaInfo: {
    title: 'Kubernetes Authentication',
    htmlAttrs: {
      lang: 'en',
    }
  },
  data: function() {
    return {
      lifetime: 2,
      clusterID: "radcluster",
      token: null,
      error: null,
    }
  },
  methods: {
    'fetchToken': function() {
      var lifetimeInHours = this.lifetime * 24 + 'h';
      var _this = this;
      this.axios.post("http://localhost/generate", {lifetime: lifetimeInHours}).then(function(response) {
        _this.token = response.token;
      }).catch(function(e) {
        if (e.response) {
          if (e.response.data.error) {
            _this.error = e.response.data.error;
            return;
          }
        }
        if (e.request) {
          _this.error = "could not connect to API";
          return;
        }
        _this.error = e
      })
    },
    'editClusterID': function(operation) {
      this.clusterID = operation.api.origElements.innerHTML
    },
    'shellSnippet': function() {
      return ''
        + 'export KUBE_CLUSTER=' + this.clusterID + '\n'
        + 'export KUBE_TOKEN="' + this.token + '"\n'
        + '\n'
        + '# Create or update your Kubernetes user.\n'
        + 'kubectl config set-credentials ${KUBE_CLUSTER} --token="${KUBE_TOKEN}"\n'
        + '\n'
        + '# Associate your Kubernetes user with an existing cluster.\n'
        + 'kubectl config set-context ${KUBE_CLUSTER} --cluster=${KUBE_CLUSTER} --user=${KUBE_CLUSTER}\n';
    },
  },
}
</script>
