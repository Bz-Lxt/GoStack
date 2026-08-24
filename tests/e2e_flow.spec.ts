import { test, expect } from '@playwright/test'

test('dashboard loads and traffic flips state wall', async ({ page, request }) => {
  const api = process.env.API_URL || 'http://127.0.0.1:28472'
  const health = await request.get(`${api}/api/v1/health`)
  expect(health.ok()).toBeTruthy()
  const body = await health.json()
  expect(body.ok).toBeTruthy()
  expect(body.data.tz).toBe('Asia/Shanghai')

  await page.goto('/')
  await expect(page.getByRole('heading', { name: /TCP 状态机翻转墙/ })).toBeVisible()
  await expect(page.getByText('MINI GOSTACK')).toBeVisible()

  await page.getByRole('button', { name: '放行流量' }).click()
  await expect(page.getByText(/流量完成|kernel dial|stack dial|match=/)).toBeVisible({ timeout: 20000 })

  await expect(page.getByText(/ESTABLISHED|FIN_WAIT|TIME_WAIT|CLOSE_WAIT|LAST_ACK|SYN_/)).toBeVisible({ timeout: 10000 })
})
