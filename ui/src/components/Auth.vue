<template>
  <div class="wrapper">
    <form v-on:submit.prevent="onSubmit">
      <h1>Access Code</h1>
      <div>
        <input type="password" name="code" required />
      </div>
      <div>
        <button type="submit">Submit</button>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  methods: {
    onSubmit: async function (event) {
      const formData = new FormData(event.target);
      const accessCode = formData.get('code');

      try {
        await this.$store.dispatch('logIn', accessCode);
      } catch (err) {
        alert('try again');
      }
    },
  },
}
</script>

<style scoped>
.wrapper {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

h1 {
  text-align: center;
}

form > div {
  margin-bottom: 1em;
}

input {
  padding: 0.5em 1em;
  font-size: 1.2em;
  border: 2px solid var(--colorPrimary);
}

button {
  padding: 0.5em 1em;
  border: 2px solid var(--colorPrimary);
  width: 100%;
  font-size: 1.2em;
}
</style>
