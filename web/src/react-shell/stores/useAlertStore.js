import { create } from 'zustand'

export const useAlertStore = create((set) => ({
  firingEvents: [],
  setFiringEvents: (events) => set({ firingEvents: events }),

  selectedRuleIds: [],
  setSelectedRuleIds: (ids) => set({ selectedRuleIds: ids }),

  filterStatus: 'firing',
  setFilterStatus: (s) => set({ filterStatus: s }),
}))
