<template>
  <div>
    <AppNav />
    <main>
      <HeroSection />
      <FeaturesSection />
      <SourcesSection />
      <ToolsSection />
      <CtaSection />
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
  title: "NezhaCyberMCP — AI-Powered CVE Intelligence",
  meta: [
    {
      name: "description",
      content: t("hero.description"),
    },
  ],
});
</script>
