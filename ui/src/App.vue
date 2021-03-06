<template>
  <div>
    <div v-if="isAuthenticated">
      <DOMObservers />

      <h1>Photos</h1>

      <div v-if="isLoading">
        <h2 class="pending">...</h2>
        <div class="masonry">
          <template>
            <Skeleton v-bind:ref="`skeleton${index}`" v-for="index in 4" v-bind:key="index" />
          </template>
        </div>
      </div>
      <div v-else v-cloak>
        <Shortcuts class="shortcuts" v-bind:buckets="buckets" />

        <div v-for="grouping in groupings" v-bind:key="grouping">
          <div v-bind:id="grouping"></div>
          <div v-for="bucket in bucketsByGroupings[grouping].buckets" v-bind:key="bucket.id">
            <Photos v-bind:bucket="bucket">
              <template v-slot="{ photo }">
                <Thumbnail v-bind:photo="photo" />
              </template>
            </Photos>
          </div>
        </div>
      </div>

      <PhotoModal v-if="modalPhoto" v-bind:photo="modalPhoto" />
    </div>
    <div v-else>
      <Auth />
    </div>
  </div>
</template>

<script>
import Vuex from 'vuex';
import Auth from './components/Auth';
import DOMObservers from './components/DOMObservers';
import PhotoModal from './components/PhotoModal';
import Photos from './components/Photos';
import Shortcuts from './components/Shortcuts.vue';
import Skeleton from './components/Skeleton.vue';
import Thumbnail from './components/Thumbnail.vue';
import store from './store';

export default {
  components: {
    Auth,
    DOMObservers,
    PhotoModal,
    Photos,
    Shortcuts,
    Skeleton,
    Thumbnail,
  },

  store,

  watch: {
    isAuthenticated: function (isAuthenticated) {
      if (isAuthenticated) {
        this.loadSkeletonData();
      }
    },
  },

  created: function () {
    this.$store.dispatch('loadInitialState');
  },

  computed: {
    ...Vuex.mapState({
      modalPhoto: state => state.modalPhoto,
      isAuthenticated: state => state.isAuthenticated,
      apiClient: state => state.apiClient,
      isLoading: state => state.isLoading,
      buckets: state => Object.values(state.bucketsByID),
    }),
    groupings: function () {
      return [...new Set((this.buckets || []).map(b => b.grouping))]
    },
    bucketsByGroupings: function () {
      return (this.buckets || []).reduce((memo, next) => {
        const existing = memo[next.grouping] || { grouping: next.grouping, buckets: [] };

        return {
          ...memo,
          [next.grouping]: {
            ...existing,
            buckets: [
              ...existing.buckets,
              next,
            ],
          },
        };
      }, {});
    },
  },

  methods: {
    loadSkeletonData: async function () {
      this.$store.commit('startLoadingDataOutline');

      const buckets = await this.apiClient('api/buckets/counts');

      this.$store.commit('loadedDataOutline', buckets);
    },
  }
};
</script>
