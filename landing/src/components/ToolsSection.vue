<template>
  <section id="tools" class="section">
    <div class="container">
      <div class="text-center section-header">
        <span class="badge">
          <i class="pi pi-wrench" />
          MCP Tools
        </span>
        <h2 class="section-title">{{ $t("tools.title") }}</h2>
        <p class="section-subtitle text-muted">{{ $t("tools.subtitle") }}</p>
        <div class="divider" />
      </div>

      <div class="tools-grid">
        <div v-for="tool in tools" :key="tool.name" class="tool-item">
          <div class="tool-item__left">
            <span class="tool-badge">
              <i :class="`pi ${tool.icon}`" />
            </span>
            <code class="tool-name">{{ tool.name }}</code>
          </div>
          <p class="tool-desc text-muted">{{ $t(`tools.items.${tool.key}`) }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
const tools = [
  { key: "get_cve", name: "get_cve", icon: "pi-search-plus" },
  { key: "search_cves", name: "search_cves", icon: "pi-filter" },
  { key: "whats_new", name: "whats_new", icon: "pi-bell" },
  { key: "filter_severity", name: "filter_severity", icon: "pi-exclamation-triangle" },
  { key: "bulk_get", name: "bulk_get", icon: "pi-list" },
  { key: "vuln_trends", name: "vuln_trends", icon: "pi-chart-line" },
  { key: "search_by_cpe", name: "search_by_cpe", icon: "pi-tag" },
  { key: "related_cves", name: "related_cves", icon: "pi-share-alt" },
];
</script>

<style scoped>
.section-header {
  margin-bottom: 56px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.section-title {
  font-size: clamp(1.8rem, 3.5vw, 2.5rem);
  font-weight: 800;
  letter-spacing: -0.02em;
}

.section-subtitle {
  font-size: 1.05rem;
  max-width: 480px;
  line-height: 1.7;
}

.tools-grid {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.tool-item {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 20px 28px;
  border-bottom: 1px solid var(--color-border);
  transition: background 0.15s;
}

.tool-item:last-child {
  border-bottom: none;
}

.tool-item:hover {
  background: rgba(99, 102, 241, 0.04);
}

.tool-item__left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  min-width: 220px;
}

.tool-badge {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-primary-light);
  font-size: 0.9rem;
  flex-shrink: 0;
}

.tool-name {
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 0.9rem;
  color: var(--color-accent);
  font-weight: 600;
}

.tool-desc {
  font-size: 0.9rem;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .tool-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 16px 20px;
  }

  .tool-item__left {
    min-width: unset;
  }
}
</style>
