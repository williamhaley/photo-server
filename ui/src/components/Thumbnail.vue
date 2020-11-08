<template>
  <img class="media pending" :data-src="src" v-on:click="showPhoto" alt="" />
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
      return `${process.env.VUE_APP_ROOT_URL}thumbnail/${this.photo.uuid}.${extension}?${params.toString()}`;
    },
    title: function () {
      return `${this.photo.name} - ${this.photo.date}`;
    },
  },

  methods: {
    showPhoto() {
      this.$store.commit('setModalPhoto', this.photo);
    }
  },
};
</script>

<style src="../../public/common.css"></style>
