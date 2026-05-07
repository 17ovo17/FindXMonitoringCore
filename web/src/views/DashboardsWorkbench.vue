<template>
  <div class="dash-page">
    <DashboardListPanel
      v-if="workbench.section.value === 'list'"
      v-model:keyword="workbench.keyword.value"
      :loading="workbench.loading.value"
      :scope-filter="workbench.scopeFilter.value"
      :error-title="workbench.errorTitle.value"
      :error-text="workbench.errorText.value"
      :permission-error="workbench.permissionError.value"
      :business-groups="workbench.businessGroups.value"
      :filtered-dashboards="workbench.filteredDashboards.value"
      :column-options="workbench.columnOptions"
      :visible-columns="workbench.visibleColumns"
      @update:selected-rows="workbench.selectedRows.value = $event"
      @set-scope="workbench.scopeFilter.value = $event"
      @refresh="workbench.loadDashboards"
      @create="workbench.openCreate"
      @templates="workbench.goTemplates"
      @batch-command="workbench.handleBatchCommand"
      @row-command="workbench.handleRowCommand"
      @open-detail="workbench.openDetail"
    />
    <DashboardDetailPanel
      v-else-if="workbench.section.value === 'detail'"
      v-model:active-dashboard-id="workbench.activeDashboardId.value"
      v-model:auto-refresh="workbench.autoRefresh.value"
      v-model:time-range="workbench.timeRange.value"
      v-model:timezone="workbench.timezone.value"
      :detail-loading="workbench.detailLoading.value"
      :detail-error="workbench.detailError.value"
      :dashboards="workbench.dashboards.value"
      :active-dashboard="workbench.activeDashboard.value"
      :variables="workbench.variables.value"
      :variable-values="workbench.variableValues"
      :panels="workbench.panels.value"
      :panel-types="workbench.panelTypes"
      @go-list="workbench.goList"
      @open-detail="workbench.openDetail"
      @refresh-detail="workbench.refreshDetail"
      @panel-editor="workbench.openPanelEditor"
      @panel-command="workbench.handlePanelCommand"
      @settings="workbench.openSettings"
      @copy-link="workbench.copyDetailLink"
      @fullscreen="workbench.toggleFullscreen"
    />
    <DashboardTemplatesPanel
      v-else
      :templates="workbench.templates.value"
      :templates-loading="workbench.templatesLoading.value"
      :templates-error="workbench.templatesError.value"
      @go-list="workbench.goList"
      @load-templates="workbench.loadTemplates"
      @preview-template="workbench.previewTemplate"
      @open-import="workbench.openImport"
    />
    <DashboardDialogs
      v-model:form-visible="workbench.formVisible.value"
      v-model:panel-drawer-visible="workbench.panelDrawerVisible.value"
      v-model:preview-visible="workbench.previewVisible.value"
      v-model:import-visible="workbench.importVisible.value"
      v-model:tag-text="workbench.tagText.value"
      v-model:variables-json="workbench.variablesJson.value"
      v-model:import-tag-text="workbench.importTagText.value"
      :saving="workbench.saving.value"
      :importing="workbench.importing.value"
      :editing-id="workbench.editingId.value"
      :form="workbench.form"
      :form-error="workbench.formError.value"
      :panel-drawer-title="workbench.panelDrawerTitle.value"
      :panel-draft-json="workbench.panelDraftJson.value"
      :preview-title="workbench.previewTitle.value"
      :preview-json="workbench.previewJson.value"
      :selected-template="workbench.selectedTemplate.value"
      :import-form="workbench.importForm"
      :import-error="workbench.importError.value"
      @save-dashboard="workbench.saveDashboard"
      @open-import="workbench.openImport"
      @submit-import="workbench.submitImport"
    />
  </div>
</template>

<script setup>
import DashboardDetailPanel from '../components/dashboard/DashboardDetailPanel.vue'
import DashboardDialogs from '../components/dashboard/DashboardDialogs.vue'
import DashboardListPanel from '../components/dashboard/DashboardListPanel.vue'
import DashboardTemplatesPanel from '../components/dashboard/DashboardTemplatesPanel.vue'
import { useDashboardWorkbench } from '../components/dashboard/useDashboardWorkbench'

const workbench = useDashboardWorkbench()
</script>

<style scoped>
.dash-page { min-height: 100%; padding: 18px; color: #25324a; background: #f5f7fb; }
</style>
