/**
 * Copyright 2018 Planet Labs Inc
 * 
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * 
 *     http://www.apache.org/licenses/LICENSE-2.0
 * 
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

new Vue({
  el: '#app',
  render: h => h(Kubehook),
  template: '<Kubehook/>',
  components: { Kubehook }
})
