<template>
  <div id="modal" v-on:click="hidePhoto">
    <Photo v-bind:photo="photo" />
  </div>
</template>

<script>
import Photo from './Photo.vue';

export default {
  props: ['photo'],

  components: {
    Photo,
  },

  mounted: function () {
    window.addEventListener('keyup', this.keyListener);
  },

  beforeDestroy() {
    window.removeEventListener('keyup', this.keyListener);
  },

  methods: {
    hidePhoto: function () {
      this.$store.commit('setModalPhoto', null);
    },
    keyListener: function (event) {
      if (event.key === 'Escape') {
        this.hidePhoto();
      }
    },
  },
};
</script>

<style scoped>
#modal {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  background-color: rgba(0, 0, 0, 0.8);
  z-index: 200;
  display: flex;
  justify-content: center;
}

#modal .media {
  margin: 1em;
}
</style>
