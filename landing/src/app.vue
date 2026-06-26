<template>
  <NuxtPage />
</template>

<script setup lang="ts">
const { init } = useTheme();

onMounted(() => {
  init();
});
</script>
