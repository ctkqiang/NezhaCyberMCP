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
  htmlAttrs: { lang: locale.value },
  title: "Installation Guide — NezhaCyberMCP",
  meta: [
    {
      name: "description",
      content: t("install.desc"),
    },
    {
      name: "author",
      content: "哪吒网络安全钟智强",
    },
    {
      name: "keywords",
      content: "NezhaCyberMCP install, Trae IDE, Claude Code, OpenAI Codex, MCP setup, 哪吒网络安全, 钟智强",
    },
    {
      property: "og:title",
      content: "Installation Guide — NezhaCyberMCP",
    },
    {
      property: "og:description",
      content: t("install.desc"),
    },
    {
      name: "viewport",
      content: "width=device-width, initial-scale=1, viewport-fit=cover",
    },
  ],
});
</script>
