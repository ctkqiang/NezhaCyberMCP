<template>
  <section class="section cta-section">
    <div class="container">
      <div class="cta-card">
        <div class="cta-card__bg" />
        <div class="cta-content">
          <h2 class="cta-title">{{ $t("cta.title") }}</h2>
          <p class="cta-desc text-muted">{{ $t("cta.desc") }}</p>
          <a
            href="https://github.com/ctkqiang/NezhaCyberMCP#readme"
            target="_blank"
            rel="noopener"
            class="btn-primary"
          >
            <i class="pi pi-book" />
            {{ $t("cta.button") }}
          </a>
        </div>

        <div class="cta-code">
          <div class="code-block">
            <div class="code-block__header">
              <span class="code-block__title">{{ $t("cta.config_title") }}</span>
              <button class="code-copy" @click="copy" :title="copied ? 'Copied!' : 'Copy'">
                <i :class="copied ? 'pi pi-check' : 'pi pi-copy'" />
              </button>
            </div>
            <pre class="code-block__body"><code><span class="t-brace">{</span>
  <span class="t-key">"mcpServers"</span><span class="t-colon">:</span> <span class="t-brace">{</span>
    <span class="t-key">"nezha-cyber"</span><span class="t-colon">:</span> <span class="t-brace">{</span>
      <span class="t-key">"command"</span><span class="t-colon">:</span> <span class="t-str">"./advisory"</span>
    <span class="t-brace">}</span>
  <span class="t-brace">}</span>
<span class="t-brace">}</span></code></pre>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
const copied = ref(false);

const configSnippet = `{
  "mcpServers": {
    "nezha-cyber": {
      "command": "./advisory"
    }
  }
}`;

async function copy() {
  await navigator.clipboard.writeText(configSnippet);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}
</script>

<style scoped>
.cta-section {
  padding-bottom: 120px;
}

.cta-card {
  position: relative;
  background: var(--color-surface);
  border: 1px solid rgba(229, 62, 62, 0.2);
  border-radius: var(--radius-xl);
  padding: 64px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
  overflow: hidden;
}

.cta-card__bg {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 30% 50%, rgba(229, 62, 62, 0.07) 0%, transparent 70%);
  pointer-events: none;
}

.cta-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
  position: relative;
  z-index: 1;
}

.cta-title {
  font-size: clamp(1.6rem, 3vw, 2.2rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.cta-desc {
  font-size: 1rem;
  line-height: 1.7;
}

.cta-code {
  position: relative;
  z-index: 1;
}

.code-block {
  background: var(--code-bg);
  border: 1px solid var(--code-border);
  border-radius: 12px;
  overflow: hidden;
}

.code-block__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--code-bar);
  border-bottom: 1px solid var(--code-border);
}

.code-block__title {
  font-size: 0.8rem;
  color: var(--code-dim);
  font-weight: 500;
}

.code-copy {
  background: none;
  border: none;
  color: var(--code-dim);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: color 0.2s, background 0.2s;
  font-size: 0.85rem;
}

.code-copy:hover {
  color: var(--code-text);
  background: var(--color-border);
}

.code-block__body {
  padding: 20px;
  font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
  font-size: 0.84rem;
  line-height: 1.85;
  color: var(--code-text);
  overflow-x: auto;
  margin: 0;
}

/* JSON syntax token colors */
.t-brace  { color: #e2b96f; }
.t-key    { color: #79c0ff; }
.t-colon  { color: var(--code-dim); }
.t-str    { color: #a5d6a7; }

@media (max-width: 1024px) {
  .cta-card {
    grid-template-columns: 1fr;
    gap: 40px;
    padding: 40px;
  }
}

@media (max-width: 480px) {
  .cta-card {
    padding: 28px 20px;
    border-radius: var(--radius-lg);
  }
}
</style>
