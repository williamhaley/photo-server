<template>
  <div>
    <div v-if="isAuthenticated">
      <DOMObservers />

      <h1>Photos</h1>

      <div v-if="loading">
        <h2 class="pending">...</h2>
        <div class="masonry">
          <template>
            <Skeleton v-bind:ref="`skeleton${index}`" v-for="index in 4" v-bind:key="index" />
          </template>
        </div>
      </div>
      <div v-else v-cloak>
        <Shortcuts class="shortcuts" v-bind:buckets="buckets" />

        <GroupedBuckets v-bind:buckets="buckets">
          <template v-slot="{ bucket }">
            <Photos v-bind:bucket="bucket">
              <template v-slot="{ photo }">
                <Thumbnail v-bind:photo="photo" />
              </template>
            </Photos>
          </template>
        </GroupedBuckets>
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
import GroupedBuckets from './components/GroupedBuckets.vue';
import PhotoModal from './components/PhotoModal';
import Photos from './components/Photos';
import Shortcuts from './components/Shortcuts.vue';
import Skeleton from './components/Skeleton.vue';
import Thumbnail from './components/Thumbnail.vue';
import { YearMonthBucket } from './common';
import store from './store';

export default {
  components: {
    Auth,
    DOMObservers,
    GroupedBuckets,
    PhotoModal,
    Photos,
    Shortcuts,
    Skeleton,
    Thumbnail,
  },

  store,

  data: function () {
    return {
      loading: true,
    };
  },

  watch: {
    isAuthenticated: function (isAuthenticated) {
      if (isAuthenticated) {
        this.loadSkeletonData();
      }
    },
  },

  computed: {
    ...Vuex.mapState({
      modalPhoto: state => state.modalPhoto,
      isAuthenticated: state => state.isAuthenticated,
      apiClient: state => state.apiClient,
    }),
  },

  methods: {
    loadSkeletonData: async function () {
      const response = await this.apiClient('api/buckets/counts');

      this.buckets = response.map(bucket => {
        return new YearMonthBucket(`${bucket.year}-${bucket.month}`, bucket.totalCount);
      });
      this.loading = false;
    },
  }
};
</script>
