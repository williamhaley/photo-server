import Vue from 'vue';
import Vuex from 'vuex';
import { YearMonthBucket } from './common';

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {
    count: 0,
    modalPhoto: null,
    isScrolling: false,
    isAuthenticated: false,
    token: null,
    apiClient: null,
    bucketsByID: {},
    isLoading: false,
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

        console.log(`fetched ${path} in ${(Date.now() - start) / 1000}s ${JSON.stringify(json)}`);

        return json;
      };
    },

    startLoadingDataOutline(state) {
      state.isLoading = true;
      state.bucketsByID = {};
    },

    loadedDataOutline(state, buckets) {
      const bucketsByID = {};

      // The initial app load returns *all* buckets, but with no photos. Only meta data to help
      // outline the view and set up later async data loading. These buckts are all sorted by the
      // API when returned, so it's important we maintain that state appropriately here.
      for (let bucketJSON of buckets) {
        // Formalized bucket that tracks various bits of state.
        const yearMonthBucket = new YearMonthBucket(`${bucketJSON.year}-${bucketJSON.month}`, bucketJSON.totalCount);

        bucketsByID[yearMonthBucket.id] = yearMonthBucket;
      }

      state.bucketsByID = bucketsByID;
      state.isLoading = false;
    },

    loadedPhotosForBucket(state, { bucketID, photos }) {
      state.bucketsByID[bucketID].appendPhotos(photos);
    },
  }
});

export default store;
