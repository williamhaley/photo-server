<template>
  <div>
    <div v-for="grouping in groupings" v-bind:key="grouping">
      <div v-bind:id="grouping"></div>
      <div v-for="bucket in bucketsByGroupings[grouping].buckets" v-bind:key="bucket.id">
        <slot v-bind:bucket="bucket" />
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: ['buckets'],

  computed: {
    groupings: function () {
      return [...new Set(this.buckets.map(b => b.grouping))]
    },

    bucketsByGroupings: function () {
      return this.buckets.reduce((memo, next) => {
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
};
</script>
