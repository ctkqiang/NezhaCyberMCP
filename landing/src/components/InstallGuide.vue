<template>
  <div class="install-guide">

    <header class="guide-hero">
      <div class="container">
        <div class="guide-hero__badge">
          <i class="pi pi-book" />
          <span>{{ $t("install.badge") }}</span>
        </div>
        <h1 class="guide-hero__title">{{ $t("install.title") }}</h1>
        <p class="guide-hero__desc text-muted">{{ $t("install.desc") }}</p>

        <div class="guide-hero__meta">
          <span class="meta-chip"><i class="pi pi-desktop" /> macOS · Linux · Windows</span>
          <span class="meta-chip"><i class="pi pi-clock" /> ~15 min</span>
          <span class="meta-chip"><i class="pi pi-user" /> Intermediate</span>
        </div>

        <nav class="toc">
          <p class="toc__label">{{ $t("install.toc_label") }}</p>
          <ol class="toc__list">
            <li v-for="item in tocItems" :key="item.id">
              <a :href="`#${item.id}`">{{ item.label }}</a>
            </li>
          </ol>
        </nav>
      </div>
    </header>

    <main class="guide-body">
      <div class="container guide-layout">
        <article class="guide-content">

          <!-- ── Section 1: Prerequisites ── -->
          <section id="prerequisites" class="guide-section">
            <div class="section-header">
              <span class="step-num">01</span>
              <div>
                <h2>{{ $t("install.s1.title") }}</h2>
                <p class="text-muted">{{ $t("install.s1.desc") }}</p>
              </div>
            </div>

            <div class="compat-grid">
              <div class="compat-card" v-for="os in osList" :key="os.name">
                <i :class="`pi ${os.icon}`" class="compat-card__icon" />
                <div>
                  <strong>{{ os.name }}</strong>
                  <p class="text-muted">{{ os.version }}</p>
                </div>
              </div>
            </div>

            <div class="req-table">
              <div class="req-row req-row--header">
                <span>{{ $t("install.s1.req_name") }}</span>
                <span>{{ $t("install.s1.req_version") }}</span>
                <span>{{ $t("install.s1.req_purpose") }}</span>
              </div>
              <div class="req-row" v-for="req in requirements" :key="req.name">
                <span class="req-name"><i :class="`pi ${req.icon}`" /> {{ req.name }}</span>
                <span class="req-version">{{ req.version }}</span>
                <span class="text-muted">{{ $t(req.purposeKey) }}</span>
              </div>
            </div>
          </section>

          <!-- ── Section 2: Trae IDE ── -->
          <section id="trae" class="guide-section">
            <div class="section-header">
              <span class="step-num">02</span>
              <div>
                <h2>{{ $t("install.s2.title") }}</h2>
                <p class="text-muted">{{ $t("install.s2.desc") }}</p>
              </div>
            </div>

            <div class="step-list">
              <div class="step" v-for="(step, i) in traeSteps" :key="i">
                <div class="step__num">{{ i + 1 }}</div>
                <div class="step__body">
                  <h4>{{ $t(step.titleKey) }}</h4>
                  <p class="text-muted">{{ $t(step.descKey) }}</p>
                  <div v-if="step.cmd" class="cmd-block">
                    <code>{{ step.cmd }}</code>
                    <button class="cmd-copy" @click="copyCmd(step.cmd)" :title="$t('install.copy')">
                      <i class="pi pi-copy" />
                    </button>
                  </div>
                  <div v-if="step.note" class="callout callout--info">
                    <i class="pi pi-info-circle" />
                    <span>{{ $t(step.note) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="verify-block">
              <div class="verify-block__label"><i class="pi pi-check-circle" /> {{ $t("install.verify") }}</div>
              <p class="text-muted">{{ $t("install.s2.verify") }}</p>
            </div>

            <div class="ref-links">
              <a href="https://www.trae.ai" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-external-link" /> trae.ai
              </a>
              <a href="https://docs.trae.ai" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-book" /> {{ $t("install.docs") }}
              </a>
            </div>
          </section>

          <!-- ── Section 3: Claude Code ── -->
          <section id="claude" class="guide-section">
            <div class="section-header">
              <span class="step-num">03</span>
              <div>
                <h2>{{ $t("install.s3.title") }}</h2>
                <p class="text-muted">{{ $t("install.s3.desc") }}</p>
              </div>
            </div>

            <div class="step-list">
              <div class="step" v-for="(step, i) in claudeSteps" :key="i">
                <div class="step__num">{{ i + 1 }}</div>
                <div class="step__body">
                  <h4>{{ $t(step.titleKey) }}</h4>
                  <p class="text-muted">{{ $t(step.descKey) }}</p>
                  <div v-if="step.cmd" class="cmd-block">
                    <code>{{ step.cmd }}</code>
                    <button class="cmd-copy" @click="copyCmd(step.cmd)" :title="$t('install.copy')">
                      <i class="pi pi-copy" />
                    </button>
                  </div>
                  <div v-if="step.note" class="callout callout--warn">
                    <i class="pi pi-exclamation-triangle" />
                    <span>{{ $t(step.note) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="verify-block">
              <div class="verify-block__label"><i class="pi pi-check-circle" /> {{ $t("install.verify") }}</div>
              <div class="cmd-block">
                <code>claude --version</code>
                <button class="cmd-copy" @click="copyCmd('claude --version')" :title="$t('install.copy')">
                  <i class="pi pi-copy" />
                </button>
              </div>
            </div>

            <div class="ref-links">
              <a href="https://claude.ai/code" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-external-link" /> claude.ai/code
              </a>
              <a href="https://docs.anthropic.com/en/docs/claude-code" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-book" /> {{ $t("install.docs") }}
              </a>
            </div>
          </section>

          <!-- ── Section 4: OpenAI Codex ── -->
          <section id="codex" class="guide-section">
            <div class="section-header">
              <span class="step-num">04</span>
              <div>
                <h2>{{ $t("install.s4.title") }}</h2>
                <p class="text-muted">{{ $t("install.s4.desc") }}</p>
              </div>
            </div>

            <div class="step-list">
              <div class="step" v-for="(step, i) in codexSteps" :key="i">
                <div class="step__num">{{ i + 1 }}</div>
                <div class="step__body">
                  <h4>{{ $t(step.titleKey) }}</h4>
                  <p class="text-muted">{{ $t(step.descKey) }}</p>
                  <div v-if="step.cmd" class="cmd-block">
                    <code>{{ step.cmd }}</code>
                    <button class="cmd-copy" @click="copyCmd(step.cmd)" :title="$t('install.copy')">
                      <i class="pi pi-copy" />
                    </button>
                  </div>
                  <div v-if="step.note" class="callout callout--warn">
                    <i class="pi pi-exclamation-triangle" />
                    <span>{{ $t(step.note) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="verify-block">
              <div class="verify-block__label"><i class="pi pi-check-circle" /> {{ $t("install.verify") }}</div>
              <div class="cmd-block">
                <code>codex --version</code>
                <button class="cmd-copy" @click="copyCmd('codex --version')" :title="$t('install.copy')">
                  <i class="pi pi-copy" />
                </button>
              </div>
            </div>

            <div class="ref-links">
              <a href="https://github.com/openai/codex" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-github" /> openai/codex
              </a>
              <a href="https://platform.openai.com/docs" target="_blank" rel="noopener" class="ref-link">
                <i class="pi pi-book" /> {{ $t("install.docs") }}
              </a>
            </div>
          </section>

          <!-- ── Section 5: NezhaCyberMCP ── -->
          <section id="nezha" class="guide-section">
            <div class="section-header">
              <span class="step-num">05</span>
              <div>
                <h2>{{ $t("install.s5.title") }}</h2>
                <p class="text-muted">{{ $t("install.s5.desc") }}</p>
              </div>
            </div>

            <div class="step-list">
              <div class="step" v-for="(step, i) in nezhaSteps" :key="i">
                <div class="step__num">{{ i + 1 }}</div>
                <div class="step__body">
                  <h4>{{ $t(step.titleKey) }}</h4>
                  <p class="text-muted">{{ $t(step.descKey) }}</p>
                  <div v-if="step.cmd" class="cmd-block">
                    <code>{{ step.cmd }}</code>
                    <button class="cmd-copy" @click="copyCmd(step.cmd)" :title="$t('install.copy')">
                      <i class="pi pi-copy" />
                    </button>
                  </div>
                  <div v-if="step.note" class="callout callout--info">
                    <i class="pi pi-info-circle" />
                    <span>{{ $t(step.note) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- MCP config highlight block -->
            <div class="config-block">
              <div class="config-block__header">
                <span>mcp.json</span>
                <button class="cmd-copy" @click="copyCmd(mcpConfig)" :title="$t('install.copy')">
                  <i class="pi pi-copy" />
                </button>
              </div>
              <pre class="config-block__body"><code><span class="t-brace">{</span>
  <span class="t-key">"mcpServers"</span><span class="t-colon">:</span> <span class="t-brace">{</span>
    <span class="t-key">"nezha-cyber"</span><span class="t-colon">:</span> <span class="t-brace">{</span>
      <span class="t-key">"command"</span><span class="t-colon">:</span> <span class="t-str">"./advisory"</span>
    <span class="t-brace">}</span>
  <span class="t-brace">}</span>
<span class="t-brace">}</span></code></pre>
            </div>

            <div class="verify-block">
              <div class="verify-block__label"><i class="pi pi-check-circle" /> {{ $t("install.verify") }}</div>
              <p class="text-muted">{{ $t("install.s5.verify") }}</p>
            </div>
          </section>

          <!-- ── Section 6: Troubleshooting ── -->
          <section id="troubleshoot" class="guide-section">
            <div class="section-header">
              <span class="step-num">06</span>
              <div>
                <h2>{{ $t("install.s6.title") }}</h2>
                <p class="text-muted">{{ $t("install.s6.desc") }}</p>
              </div>
            </div>

            <div class="trouble-list">
              <details class="trouble-item" v-for="(item, i) in troubleItems" :key="i">
                <summary class="trouble-item__q">
                  <i class="pi pi-question-circle" />
                  {{ $t(item.qKey) }}
                </summary>
                <div class="trouble-item__a">
                  <p class="text-muted">{{ $t(item.aKey) }}</p>
                  <div v-if="item.cmd" class="cmd-block">
                    <code>{{ item.cmd }}</code>
                    <button class="cmd-copy" @click="copyCmd(item.cmd)" :title="$t('install.copy')">
                      <i class="pi pi-copy" />
                    </button>
                  </div>
                </div>
              </details>
            </div>
          </section>

        </article>

        <!-- Sticky sidebar TOC (desktop) -->
        <aside class="guide-sidebar">
          <div class="sidebar-toc">
            <p class="sidebar-toc__label">{{ $t("install.toc_label") }}</p>
            <ul>
              <li v-for="item in tocItems" :key="item.id">
                <a :href="`#${item.id}`" :class="{ active: activeSection === item.id }">
                  <span class="toc-num">{{ item.num }}</span>
                  {{ item.label }}
                </a>
              </li>
            </ul>
          </div>
        </aside>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n();

const osList = [
  { name: "macOS", icon: "pi-apple",   version: "12 Monterey+" },
  { name: "Linux", icon: "pi-server",  version: "Ubuntu 20.04+ / Debian 11+" },
  { name: "Windows", icon: "pi-microsoft", version: "Windows 10 / 11 (WSL2 recommended)" },
];

const requirements = [
  { name: "Go",   icon: "pi-code",     version: "1.22+",  purposeKey: "install.s1.go_purpose" },
  { name: "Git",  icon: "pi-github",   version: "2.x",    purposeKey: "install.s1.git_purpose" },
  { name: "Node", icon: "pi-box",      version: "18 LTS+", purposeKey: "install.s1.node_purpose" },
  { name: "npm",  icon: "pi-box",      version: "9+",     purposeKey: "install.s1.npm_purpose" },
];

const traeSteps = [
  { titleKey: "install.s2.step1_title", descKey: "install.s2.step1_desc", cmd: "curl -fsSL https://www.trae.ai/install.sh | sh", note: null },
  { titleKey: "install.s2.step2_title", descKey: "install.s2.step2_desc", cmd: null, note: "install.s2.step2_note" },
  { titleKey: "install.s2.step3_title", descKey: "install.s2.step3_desc", cmd: null, note: null },
];

const claudeSteps = [
  { titleKey: "install.s3.step1_title", descKey: "install.s3.step1_desc", cmd: "npm install -g @anthropic-ai/claude-code", note: null },
  { titleKey: "install.s3.step2_title", descKey: "install.s3.step2_desc", cmd: "claude auth login", note: null },
  { titleKey: "install.s3.step3_title", descKey: "install.s3.step3_desc", cmd: null, note: "install.s3.step3_note" },
];

const codexSteps = [
  { titleKey: "install.s4.step1_title", descKey: "install.s4.step1_desc", cmd: "npm install -g @openai/codex", note: null },
  { titleKey: "install.s4.step2_title", descKey: "install.s4.step2_desc", cmd: "export OPENAI_API_KEY=sk-...", note: "install.s4.step2_note" },
  { titleKey: "install.s4.step3_title", descKey: "install.s4.step3_desc", cmd: "codex", note: null },
];

const nezhaSteps = [
  { titleKey: "install.s5.step1_title", descKey: "install.s5.step1_desc", cmd: "git clone https://github.com/ctkqiang/NezhaCyberMCP.git && cd NezhaCyberMCP", note: null },
  { titleKey: "install.s5.step2_title", descKey: "install.s5.step2_desc", cmd: "make build", note: null },
  { titleKey: "install.s5.step3_title", descKey: "install.s5.step3_desc", cmd: null, note: "install.s5.step3_note" },
  { titleKey: "install.s5.step4_title", descKey: "install.s5.step4_desc", cmd: "make run", note: null },
];

const troubleItems = [
  { qKey: "install.s6.q1", aKey: "install.s6.a1", cmd: "go version" },
  { qKey: "install.s6.q2", aKey: "install.s6.a2", cmd: "chmod +x ./advisory" },
  { qKey: "install.s6.q3", aKey: "install.s6.a3", cmd: null },
  { qKey: "install.s6.q4", aKey: "install.s6.a4", cmd: "claude mcp add nezha-cyber ./advisory" },
];

const mcpConfig = `{
  "mcpServers": {
    "nezha-cyber": {
      "command": "./advisory"
    }
  }
}`;

const tocItems = computed(() => [
  { id: "prerequisites", num: "01", label: t("install.s1.title") },
  { id: "trae",          num: "02", label: t("install.s2.title") },
  { id: "claude",        num: "03", label: t("install.s3.title") },
  { id: "codex",         num: "04", label: t("install.s4.title") },
  { id: "nezha",         num: "05", label: t("install.s5.title") },
  { id: "troubleshoot",  num: "06", label: t("install.s6.title") },
]);

const activeSection = ref("prerequisites");

onMounted(() => {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((e) => {
        if (e.isIntersecting) activeSection.value = e.target.id;
      });
    },
    { rootMargin: "-30% 0px -60% 0px" }
  );
  document.querySelectorAll(".guide-section").forEach((el) => observer.observe(el));
});

async function copyCmd(text: string) {
  await navigator.clipboard.writeText(text);
}
</script>

<style scoped>
/* ── Layout ── */
.install-guide {
  min-height: 100vh;
  background: var(--color-bg);
}

/* ── Hero ── */
.guide-hero {
  padding: 120px 0 64px;
  border-bottom: 1px solid var(--color-border);
  background: linear-gradient(180deg, var(--color-surface-2) 0%, var(--color-bg) 100%);
}

.guide-hero__badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 5px 14px;
  border-radius: 999px;
  border: 1px solid rgba(229, 62, 62, 0.3);
  background: rgba(229, 62, 62, 0.08);
  color: var(--color-primary-light);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  margin-bottom: 20px;
}

.guide-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 800;
  line-height: 1.15;
  margin-bottom: 16px;
  background: linear-gradient(135deg, var(--color-text) 0%, var(--color-primary-light) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.guide-hero__desc {
  font-size: 1.05rem;
  max-width: 640px;
  margin-bottom: 28px;
  line-height: 1.7;
}

.guide-hero__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 36px;
}

.meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 8px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  font-size: 0.82rem;
  color: var(--color-text-muted);
}

.meta-chip .pi { color: var(--color-primary-light); }

/* ── TOC (hero) ── */
.toc {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px 24px;
  max-width: 480px;
}

.toc__label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-primary-light);
  margin-bottom: 12px;
}

.toc__list {
  list-style: none;
  counter-reset: toc;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.toc__list li a {
  font-size: 0.88rem;
  color: var(--color-text-muted);
  transition: color 0.15s;
}

.toc__list li a:hover { color: var(--color-primary-light); }

/* ── Body layout ── */
.guide-body { padding: 64px 0 120px; }

.guide-layout {
  display: grid;
  grid-template-columns: 1fr 240px;
  gap: 64px;
  align-items: start;
}

/* ── Section ── */
.guide-section {
  padding-bottom: 72px;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 72px;
}

.guide-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  margin-bottom: 36px;
}

.step-num {
  font-size: 2.5rem;
  font-weight: 900;
  line-height: 1;
  color: var(--color-primary);
  opacity: 0.25;
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
}

.section-header h2 {
  font-size: 1.6rem;
  font-weight: 700;
  margin-bottom: 6px;
}

/* ── Compat grid ── */
.compat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 28px;
}

.compat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
}

.compat-card__icon {
  font-size: 1.4rem;
  color: var(--color-primary-light);
}

.compat-card strong { font-size: 0.9rem; display: block; }
.compat-card p { font-size: 0.78rem; margin-top: 2px; }

/* ── Requirements table ── */
.req-table {
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.req-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 2fr;
  gap: 16px;
  padding: 12px 20px;
  font-size: 0.86rem;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
}

.req-row:last-child { border-bottom: none; }

.req-row--header {
  background: var(--color-surface-2);
  font-weight: 700;
  font-size: 0.75rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.req-name { display: flex; align-items: center; gap: 8px; font-weight: 600; }
.req-name .pi { color: var(--color-primary-light); }

.req-version {
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.82rem;
  color: var(--color-primary-light);
  background: rgba(229, 62, 62, 0.08);
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
}

/* ── Steps ── */
.step-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  position: relative;
}

.step-list::before {
  content: "";
  position: absolute;
  left: 19px;
  top: 40px;
  bottom: 40px;
  width: 2px;
  background: linear-gradient(180deg, var(--color-primary) 0%, transparent 100%);
  opacity: 0.2;
}

.step {
  display: flex;
  gap: 20px;
  padding: 24px 0;
}

.step__num {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  color: #fff;
  font-size: 0.85rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(229, 62, 62, 0.35);
  position: relative;
  z-index: 1;
}

.step__body { flex: 1; padding-top: 8px; }
.step__body h4 { font-size: 1rem; font-weight: 600; margin-bottom: 6px; }
.step__body p  { font-size: 0.88rem; line-height: 1.65; margin-bottom: 12px; }

/* ── Command block ── */
.cmd-block {
  display: flex;
  align-items: center;
  gap: 0;
  background: var(--terminal-bg);
  border: 1px solid var(--terminal-border);
  border-radius: 8px;
  overflow: hidden;
  margin-top: 10px;
}

.cmd-block code {
  flex: 1;
  padding: 11px 16px;
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.83rem;
  color: var(--terminal-text);
  white-space: pre-wrap;
  word-break: break-all;
}

.cmd-copy {
  padding: 11px 14px;
  background: transparent;
  border: none;
  border-left: 1px solid var(--terminal-border);
  color: var(--color-text-dim);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
  flex-shrink: 0;
}

.cmd-copy:hover {
  color: var(--color-primary-light);
  background: rgba(229, 62, 62, 0.08);
}

/* ── Callouts ── */
.callout {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.84rem;
  line-height: 1.6;
  margin-top: 10px;
}

.callout--info {
  background: rgba(229, 62, 62, 0.06);
  border: 1px solid rgba(229, 62, 62, 0.18);
  color: var(--color-text-muted);
}

.callout--info .pi { color: var(--color-primary-light); margin-top: 2px; flex-shrink: 0; }

.callout--warn {
  background: rgba(246, 173, 85, 0.08);
  border: 1px solid rgba(246, 173, 85, 0.25);
  color: var(--color-text-muted);
}

.callout--warn .pi { color: var(--color-accent); margin-top: 2px; flex-shrink: 0; }

/* ── Verify block ── */
.verify-block {
  margin-top: 28px;
  padding: 20px 24px;
  background: rgba(104, 211, 145, 0.06);
  border: 1px solid rgba(104, 211, 145, 0.2);
  border-radius: 10px;
}

.verify-block__label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #68d391;
  margin-bottom: 8px;
}

/* ── Ref links ── */
.ref-links {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 20px;
}

.ref-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--color-border-2);
  font-size: 0.82rem;
  color: var(--color-text-muted);
  transition: border-color 0.15s, color 0.15s;
}

.ref-link:hover {
  border-color: rgba(229, 62, 62, 0.4);
  color: var(--color-primary-light);
}

/* ── Config block ── */
.config-block {
  margin-top: 28px;
  border: 1px solid var(--code-border);
  border-radius: 10px;
  overflow: hidden;
}

.config-block__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: var(--code-bar);
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.78rem;
  color: var(--code-dim);
}

.config-block__body {
  background: var(--code-bg);
  padding: 20px;
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.84rem;
  line-height: 1.85;
  color: var(--code-text);
  margin: 0;
  overflow-x: auto;
}

.t-brace  { color: #e2b96f; }
.t-key    { color: #79c0ff; }
.t-colon  { color: var(--code-dim); }
.t-str    { color: #a5d6a7; }

/* ── Troubleshooting ── */
.trouble-list { display: flex; flex-direction: column; gap: 12px; }

.trouble-item {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.15s;
}

.trouble-item[open] { border-color: rgba(229, 62, 62, 0.3); }

.trouble-item__q {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  font-size: 0.92rem;
  font-weight: 600;
  cursor: pointer;
  list-style: none;
  color: var(--color-text);
  transition: color 0.15s;
}

.trouble-item__q::-webkit-details-marker { display: none; }
.trouble-item__q .pi { color: var(--color-primary-light); flex-shrink: 0; }
.trouble-item[open] .trouble-item__q { color: var(--color-primary-light); }

.trouble-item__a {
  padding: 0 20px 20px 44px;
  font-size: 0.88rem;
}

/* ── Sidebar TOC ── */
.guide-sidebar { position: sticky; top: 88px; }

.sidebar-toc {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.sidebar-toc__label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-primary-light);
  margin-bottom: 14px;
}

.sidebar-toc ul { list-style: none; display: flex; flex-direction: column; gap: 4px; }

.sidebar-toc a {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 7px;
  font-size: 0.82rem;
  color: var(--color-text-muted);
  transition: background 0.15s, color 0.15s;
}

.sidebar-toc a:hover,
.sidebar-toc a.active {
  background: rgba(229, 62, 62, 0.08);
  color: var(--color-primary-light);
}

.toc-num {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--color-primary);
  opacity: 0.6;
  font-variant-numeric: tabular-nums;
  min-width: 20px;
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .guide-layout { grid-template-columns: 1fr; }
  .guide-sidebar { display: none; }
  .compat-grid { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .guide-hero { padding: 100px 0 48px; }
  .req-row { grid-template-columns: 1fr 1fr; }
  .req-row > span:last-child { grid-column: 1 / -1; }
}
</style>
