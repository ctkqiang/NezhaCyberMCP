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
    {
      name: "author",
      content: "哪吒网络安全钟智强",
    },
    {
      name: "keywords",
      content: "CVE, MCP, AI security, vulnerability intelligence, NezhaCyberMCP, 哪吒网络安全, 钟智强, cybersecurity",
    },
    {
      property: "og:title",
      content: "NezhaCyberMCP — AI-Powered CVE Intelligence",
    },
    {
      property: "og:description",
      content: t("hero.description"),
    },
    {
      property: "og:type",
      content: "website",
    },
    {
      name: "twitter:card",
      content: "summary_large_image",
    },
    {
      name: "twitter:title",
      content: "NezhaCyberMCP — AI-Powered CVE Intelligence",
    },
    {
      name: "twitter:description",
      content: t("hero.description"),
    },
    {
      name: "viewport",
      content: "width=device-width, initial-scale=1, viewport-fit=cover",
    },
  ],
});
</script>
