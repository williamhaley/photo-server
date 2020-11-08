import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {
    count: 0,
    modalPhoto: null,
    isScrolling: false,
    isAuthenticated: false,
    token: null,
    apiClient: null,
  },
  mutations: {
    setModalPhoto(state, modalPhoto) {
      if (modalPhoto) {
        document.body.style.overflowY = 'hidden';
      } else {
        document.body.style.overflowY = '';
      }
      state.modalPhoto = modalPhoto;
    },

    setIsScrolling(state) {
      state.isScrolling = true;
    },
    setScrollingSettled(state) {
      state.isScrolling = false;
    },

    setAuthInfo(state, token) {
      state.isAuthenticated = true;
      state.token = token;
      state.apiClient = async (path, opts) => {
        const start = Date.now();

        const url = `${process.env.VUE_APP_ROOT_URL}${path}`;
        const headers = new Headers();
        headers.set('Authorization', token);
        headers.set('Content-Type', 'application/json');

        const res = await fetch(url, {
          ...opts,
          headers,
        });
        const json = await res.json();

        console.log(`fetched ${path} in ${(Date.now() - start) / 1000}s`);

        return json;
      };
    }
  }
});

export default store;
