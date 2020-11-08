<template>
  <div>
    <h2>{{ this.bucket.heading }}</h2>
    <div class="masonry">
      <div class="masonry-brick" v-for="photo in this.photos" v-bind:key="'photo' + photo.uuid + photo.name">
        <slot v-bind:photo="photo" />
      </div>

      <template v-if="!this.isDone">
        <Skeleton v-bind:ref="`skeleton${index}`" v-for="index in this.numSkeletons" v-bind:key="index" />
      </template>
    </div>
  </div>
</template>

<script>
import Vuex from 'vuex';
import Skeleton from './Skeleton';

function getElementOnScreenAmount($el) {
  const bounds = $el.getBoundingClientRect();

  return Math.max(0, Math.min(window.innerHeight, bounds.height + bounds.y) - Math.max(0, bounds.y))
}

export default {
  props: ['bucket'],

  data: function () {
    return {
      numSkeletons: 4,
      photos: [],
      totalPhotosLoaded: 0,
      isDone: false,
      endCursor: '',
      loading: false,
    };
  },

  components: {
    Skeleton,
  },

  watch: {
    photos: function () {
      // Check if we should load more after we just loaded. Because if the user
      // does nothing, but skeletons are on screen, how else would we know?
      setTimeout(() => {
        this.maybeLoadMore('idle after load more');
        this.maybeLoadImages('idle after load more');
      }, 100);
    }
  },

  created: function () {
    // This is invoked the moment scrolling starts and the moment is stops. Not
    // for intermittent scrolls or other page behavior. Track this watcher so
    // we can unwatch it when we're done loading.
    this.isScrollingUnwatcher = this.$watch('isScrolling', (isScrolling) => {
      if (isScrolling) {
        return;
      }

      if (this.isDone && this.totalPhotosLoaded == this.photos.length && this.isScrollingUnwatcher) {
        console.log(`${this.bucket.id} remove scrolling watcher`);
        this.isScrollingUnwatcher();
      }

      // Scrolling stopped. We know we have focus.
      this.maybeLoadMore('scroll idle');
      this.maybeLoadImages('scroll idle');
    });

    // This is invoked whenever the skeletons intersect the viewport, which is
    // complimentary to the scroll observers. While the scroll observer checks
    // at the start or end of scrolling this can check during continuous scroll
    // that a view is now on screen.
    this.intersectionObserver = new IntersectionObserver((entries) => {
      let anySkeletons = false;

      for (let entry of entries) {
        if (!entry.isIntersecting) {
          return;
        }

        const componentTag = entry.target.__vue__.$options._componentTag;

        // If both conditions were satisfied we can short-circuit.
        if (anySkeletons) {
          break;
        }

        if (!anySkeletons) {
          if (componentTag === 'Skeleton') {
            anySkeletons = true;
          }
        }
      }

      if (anySkeletons) {
        setTimeout(() => {
          this.maybeLoadMore('intersection idle');
        }, 200);
      }
    });
  },

  mounted: function () {
    // The component mounted.
    this.maybeLoadMore('mounted');
    this.maybeLoadImages('mounted');

    // TODO WFH Photo images aren't properly scaled.

    // Observe the skeletons for intersection. Observing intersection on these
    // seems reasonable since they're finite and go in and out of view.
    [...new Array(this.numSkeletons)].forEach((_, index) => {
      const $skeleton = this.$refs[`skeleton${index + 1}`][0].$el;
      this.intersectionObserver.observe($skeleton);
    });

    // TODO WFH This is aggressive, but I don't really have a better solution
    // yet. Unlike the stateful scroll observer this can catch flings and fluid
    // motion. It isn't waiting for scrolling to stop or start. This is crucial
    // for mobile until I find a better alternative. Intersection observers are
    // kind of messy, but that may be a better route.
    // TODO WFH Could I leverage the stateful observers for isScrolling to do
    // this same behavior? Do an interval check from true until false?
    window.addEventListener('scroll', () => {
      setTimeout(() => {
        this.maybeLoadImages('scroll check (aggressive)');
      }, 150);
    });
  },

  computed: {
    ...Vuex.mapState({
      isScrolling: state => state.isScrolling,
      apiClient: state => state.apiClient,
    }),
  },

  methods: {
    isOnScreen: function () {
      const overlap = getElementOnScreenAmount(this.$el);

      if (overlap > 0) {
        console.log(`${this.bucket.id} is on screen ${overlap}`);
      }

      return overlap > 0;
    },

    maybeLoadImages: function (reason) {
      if (!this.isOnScreen()) {
        return;
      }

      const $images = this.$el.querySelectorAll('img.pending');
      let numLoaded = 0;
      for (let $img of $images) {
        if (getElementOnScreenAmount($img) > .05) {
          numLoaded++;
          $img.src = $img.dataset.src;
          $img.classList.remove('pending');
        }
      }

      this.totalPhotosLoaded += numLoaded;

      console.log(`${this.bucket.id} loaded ${numLoaded} images in viewport [${reason}]`);
    },

    maybeLoadMore: function (reason) {
      if (this.isDone) {
        return;
      }

      if (!this.isOnScreen()) {
        return;
      }

      const skeletonOverlap = [...new Array(this.numSkeletons)].map((_, index) => {
        const $skeleton = this.$refs[`skeleton${index + 1}`][0].$el;
        return getElementOnScreenAmount($skeleton);
      }).find(overlap => {
        return overlap > 0;
      });

      if (!skeletonOverlap || skeletonOverlap <= 0) {
        return;
      }

      console.log(`${this.bucket.id} has skeletons focused ${skeletonOverlap} [${reason}]`);

      this.loadMore();
    },

    loadMore: async function () {
      if (this.isDone) {
        console.log(`${this.bucket.id} done loading`);
        return;
      }

      if (this.loading) {
        console.log(`${this.bucket.id} already loading. Wait for it...`);
        return;
      }

      this.loading = true;

      const cursor = encodeURIComponent(this.endCursor);

      const params = new URLSearchParams();
      if (cursor) {
        params.append('after', cursor);
      }
      const response = await this.apiClient(`api/buckets/${this.bucket.id}?${params.toString()}`);

      response.photosConnection.edges.forEach(edge => {
        this.$set(this.photos, this.photos.length, edge.node);
      });

      this.endCursor = response.photosConnection.pageInfo.endCursor;
      if (!response.photosConnection.pageInfo.hasNextPage) {
        this.isDone = true;
      }
      console.log(`${this.bucket.id} loaded ${response.photosConnection.edges.length}`);
      console.log(`${this.bucket.id} new endcursor: ${this.endCursor}`);

      this.loading = false;
    },
  },
};
</script>
