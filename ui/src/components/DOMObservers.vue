<template>
  <span></span>
</template>

<script>
function debounce(callback, wait) {
  let timeout;
  return (...args) => {
    const context = this;
    clearTimeout(timeout);
    timeout = setTimeout(() => callback.apply(context, args), wait);
  };
}

export default {
  mounted: function () {
    // Note that this will fire just on DOM load, which could be confusing.
    // Some app behavior is based around that assumption. Search "isScrolling"
    // and keep that in mind.
    window.addEventListener('scroll', () => {
      // TODO WFH Hopefully this is cached on set() so it's not undue strain. I
      // know observers aren't re-updating, but still, seems excessive to
      // constantly set this if it may be inefficient.
      this.$store.commit('setIsScrolling');
    });

    // The timeout here has a global impact. Think about this. This is kept to
    // a relatively short time period so that the app can respond meaningfully.
    // If a component needs a longer timeout it must handle that itself. Also
    // note that scrolling carries different inertia on mobile and a fling
    // must come to an absolute stop before we decide scroll is settled.
    window.addEventListener('scroll', debounce(() => {
      console.log('scrolling settled');
      this.$store.commit('setScrollingSettled');
    }, 150));

    // TODO Don't just co-opt the scroll vars, even if the behavior is almost
    // identical between the two events.
    window.addEventListener('resize', () => {
      this.$store.commit('setIsScrolling');
    });
    window.addEventListener('resize', debounce(() => {
      console.log('resize settled');
      this.$store.commit('setScrollingSettled');
    }, 100));
  },
};
</script>
