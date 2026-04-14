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
    key: 'create-quiz',
    label: 'Создать квиз',
    icon: 'pi pi-plus',
    disabled: true,
  },
  {
    key: 'my-quizzes',
    label: 'Мои квизы',
    icon: 'pi pi-list',
    disabled: true,
  },
  {
    key: 'launch-quiz',
    label: 'Запуск квиза',
    icon: 'pi pi-send',
    disabled: true,
  },
]
