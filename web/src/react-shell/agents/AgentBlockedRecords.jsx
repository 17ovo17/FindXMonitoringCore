import React from 'react'
import { ConfigRolloutRecords } from './ConfigRolloutRecords.jsx'
import { AgentTaskRecords } from './AgentTaskRecords.jsx'
import { HostLifecycleTaskActions } from './HostLifecycleTaskActions.jsx'
import { InstallExecutionRecords } from './InstallExecutionRecords.jsx'
import { InstallPlanRecords } from './InstallPlanRecords.jsx'

export function AgentBlockedRecords({ targetIds = [], agentIds = [], targetCount = 0 }) {
  return (
    <div>
      <HostLifecycleTaskActions targetIds={targetIds} agentIds={agentIds} targetCount={targetCount} />
      <InstallPlanRecords />
      <InstallExecutionRecords />
      <ConfigRolloutRecords />
      <AgentTaskRecords />
    </div>
  )
}
