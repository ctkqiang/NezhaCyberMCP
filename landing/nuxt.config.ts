export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: { enabled: false },
  srcDir: "src/",

  modules: ["@primevue/nuxt-module", "@nuxtjs/i18n"],

  primevue: {
    options: {
      theme: "none",
      ripple: true,
    },
    autoImport: true,
  },

  i18n: {
    locales: [
      { code: "en", language: "en-US", name: "English", file: "en.json" },
      { code: "zh", language: "zh-CN", name: "中文", file: "zh.json" },
    ],
    defaultLocale: "en",
    langDir: "locales/",
    strategy: "prefix_except_default",
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: "i18n_redirected",
      redirectOn: "root",
    },
  },

  css: ["primeicons/primeicons.css", "~/assets/css/main.css"],

  app: {
    head: {
      charset: "utf-8",
      viewport: "width=device-width, initial-scale=1",
      title: "NezhaCyberMCP — AI-Powered CVE Intelligence",
      meta: [
        {
          name: "description",
          content:
            "NezhaCyberMCP is an MCP server that delivers real-time CVE intelligence from CIRCL, GitHub Advisory, and MyCERT directly to your AI assistant.",
        },
        { name: "keywords", content: "CVE, MCP, cybersecurity, vulnerability, AI, LLM, CIRCL, GitHub Advisory" },
        { name: "author", content: "ctkqiang" },
        { property: "og:type", content: "website" },
        { property: "og:title", content: "NezhaCyberMCP — AI-Powered CVE Intelligence" },
        {
          property: "og:description",
          content: "Real-time CVE intelligence for your AI assistant via the Model Context Protocol.",
        },
        { property: "og:image", content: "/logo.png" },
        { name: "twitter:card", content: "summary_large_image" },
        { name: "twitter:title", content: "NezhaCyberMCP" },
        {
          name: "twitter:description",
          content: "Real-time CVE intelligence for your AI assistant via the Model Context Protocol.",
        },
        { name: "twitter:image", content: "/logo.png" },
      ],
      link: [
        { rel: "icon", type: "image/png", href: "/logo.png" },
        { rel: "preconnect", href: "https://fonts.googleapis.com" },
        { rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: "" },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=Noto+Sans+SC:wght@300;400;500;700&display=swap",
        },
      ],
    },
  },

  nitro: {
    compressPublicAssets: true,
  },
});
