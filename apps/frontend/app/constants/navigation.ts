export type DashboardNavItem = {
  key: string
  label: string
  icon: string
  to: string
}

export const dashboardNavigationItems: DashboardNavItem[] = [
  {
    key: 'quizzes',
    label: 'Мои квизы',
    icon: 'pi pi-list',
    to: '/quizzes',
  },
  {
    key: 'quiz-editor',
    label: 'Редактор квиза',
    icon: 'pi pi-file-edit',
    to: '/quizzes/editor',
  },
]
