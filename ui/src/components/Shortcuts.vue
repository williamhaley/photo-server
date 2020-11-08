<template>
  <transition name="fade">
    <div class="shortcuts" v-if="this.isVisible">
      <div v-for="grouping in groupings" v-bind:key="grouping">
        <a v-bind:href="'#' + grouping">{{grouping}}</a>
      </div>
    </div>
  </transition>
</template>

<script>
import Vuex from 'vuex';

export default {
  props: ['buckets'],

  data: function () {
    return {
      isVisible: false,
    }
  },

  watch: {
    isScrolling: function (isScrolling) {
      if (isScrolling) {
        this.isVisible = true;

        if (this.timer) {
          clearTimeout(this.timer);
          this.timer = null;
        }
        return;
      }

      this.timer = setTimeout(() => {
        this.isVisible = false;
        this.timer = null;
      }, 3000);
    },
  },

  computed: {
    ...Vuex.mapState({
      isScrolling: state => state.isScrolling,
    }),
    groupings: function () {
      const maxShortcuts = 15;
      const allGroupings = [...new Set(this.buckets.map(b => b.grouping))];
      const distribution = Math.ceil(allGroupings.length / (maxShortcuts - 2));

      return allGroupings.filter((grouping, index) => {
        return index === 0 || index === allGroupings.length - 1 || index % distribution === 0;
      });
    },
  },
};
</script>

<style scoped>
.shortcuts {
  background-color: #fff;
  padding: 1em;
  position: fixed;
  z-index: 150;
  right: 0;
  top: 0;
  bottom: 0;
  width: 3em;
  flex-direction: column;
  justify-content: space-between;
  display: flex;
  border-left: 2px solid var(--colorPrimary);
}
.fade-enter-active, .fade-leave-active {
  transition: opacity .5s;
}
.fade-enter, .fade-leave-to {
  opacity: 0;
}
</style>
