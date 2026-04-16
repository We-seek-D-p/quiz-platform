import type { RouteRecordName } from 'vue-router'

export type DashboardNavigationItem = {
  key: string
  label: string
  icon: string
  routeName?: RouteRecordName
  disabled?: boolean
}

export const dashboardNavigationItems: DashboardNavigationItem[] = [
  {
    key: 'dashboard',
    label: 'Дашборд',
    icon: 'pi pi-home',
    routeName: 'dashboard',
  },
  {
    key: 'my-quizzes',
    label: 'Мои квизы',
    icon: 'pi pi-list',
    routeName: 'quiz-list',
  },
  {
    key: 'find-quiz',
    label: 'Найти квиз',
    icon: 'pi pi-search',
    disabled: true,
  },
]
