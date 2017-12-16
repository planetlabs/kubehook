import axios from 'axios'

import Vue from 'vue'
import VueAxios from 'vue-axios'
import Meta from 'vue-meta'
import BootstrapVue from 'bootstrap-vue'
import VueHighlightJS from 'vue-highlightjs'
import VueSliderComponent from 'vue-slider-component'

import Kubehook from './kubehook.vue'

import 'bootswatch/dist/united/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import 'highlight.js/styles/github.css'

Vue.use(VueAxios, axios);
Vue.use(Meta);
Vue.use(BootstrapVue);
Vue.use(VueHighlightJS)

Vue.component('v-slider', VueSliderComponent);
Vue.component('v-editable-code', {
  template: `<code contenteditable="true" @input="$emit('update:content', $event.target.innerText)"></code>`,
  props: ['content'],
  mounted: function () {
    this.$el.innerText = this.content;
  },
  watch: {
    content: function () {
      this.$el.innerText = this.content;
    }
  }
});

new Vue({
  el: '#app',
  render: h => h(Kubehook),
  template: '<Kubehook/>',
  components: { Kubehook }
})
