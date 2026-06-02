<template>
  <div class="dashboard">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>仪表盘</h2>
      <span class="subtitle">仪表盘</span>
    </div>
  </div>
    
</template>

<script setup>

</script>

<style scoped>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.page-header .subtitle {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
  display: block;
}

/* 统计卡片行 */
.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  cursor: default;
  transition: transform 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cluster-icon { background: #e1f0ff; color: #409eff; }
.node-icon { background: #e6f7e6; color: #67c23a; }
.ns-icon { background: #fff3e0; color: #e6a23c; }
.pod-icon { background: #f0e6ff; color: #9b59b6; }
.svc-icon { background: #e6f7ff; color: #00bcd4; }
.pv-icon { background: #fff8e1; color: #ff9800; }
.warn-icon { background: #ffe6e6; color: #f56c6c; }

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-total {
  font-size: 14px;
  font-weight: 400;
  color: #909399;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

/* 图表行 */
.chart-row {
  margin-bottom: 16px;
}

.chart-card {
  height: 100%;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* Pod 状态分布 */
.pod-status-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}

.pod-status-item {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1 1 140px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.pod-status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.pod-status-label {
  font-size: 13px;
  color: #606266;
  flex: 1;
}

.pod-status-count {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
}

/* 工作负载 */
.workload-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}

.workload-item {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1 1 140px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.workload-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.workload-info {
  display: flex;
  flex-direction: column;
}

.workload-count {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.workload-name {
  font-size: 12px;
  color: #909399;
}

/* 资源使用率 */
.resource-bars {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 8px 0;
}

.resource-bar-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.resource-bar-header {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #606266;
}

.usage-low { color: #67c23a; font-weight: 600; }
.usage-mid { color: #e6a23c; font-weight: 600; }
.usage-high { color: #f56c6c; font-weight: 600; }

.resource-bar-detail {
  font-size: 12px;
  color: #909399;
  text-align: right;
}

/* 集群列表 */
.cluster-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cluster-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.cluster-item-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.cluster-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.cluster-item-right {
  display: flex;
  gap: 16px;
}

.cluster-meta {
  font-size: 12px;
  color: #909399;
}

.cluster-meta.version {
  font-family: 'Courier New', monospace;
}

.resource-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.usage-val {
  font-weight: 600;
  color: #303133;
}

.resource-limits {
  font-size: 11px;
  color: #909399;
}
</style>
