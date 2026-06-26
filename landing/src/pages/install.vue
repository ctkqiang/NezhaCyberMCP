<template>
  <div>
    <AppNav />
    <main>
      <InstallGuide />
    </main>
    <AppFooter />
  </div>
</template>

<script setup lang="ts">
const { t, locale } = useI18n();

useHead({
  htmlAttrs: {
    lang: locale.value,
  },
  title: "Installation Guide — NezhaCyberMCP",
  meta: [
    {
      name: "description",
      content: t("install.desc"),
    },
  ],
});
</script>
