<template>
  <img class="media" v-bind:src="src" v-bind:title="title" v-bind:alt="title" />
</template>

<script>
export default {
  props: ['photo'],

  computed: {
    src: function () {
      const extension = this.photo.name.split('.').pop();
      // TODO WFH A service worker would be nice...
      const params = new URLSearchParams();
      params.append('token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTE3OTg2NTQsImZvbyI6ImJhciJ9.wTqzNEBGHmLGyfrB8XA5aPS4vo3Km42FzC6HVw0cPeQ');
      return `${process.env.VUE_APP_ROOT_URL}full/${this.photo.uuid}.${extension}?${params.toString()}`;
    },
    title: function () {
      return `${this.photo.name} - ${this.photo.date}`;
    },
  },
};
</script>

<style scoped src="../../public/common.css"></style>
