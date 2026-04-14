import { Capacitor } from '@capacitor/core';
import { LocalNotifications } from '@capacitor/local-notifications';

const REMINDER_NOTIFICATION_ID = 10001;
const REMINDER_CHANNEL_ID = 'daily-review-reminders';
const REMINDER_ENABLED_KEY = 'daily-review-reminder-enabled';
const REMINDER_TIME_KEY = 'daily-review-reminder-time';

export interface ReminderPayload {
  focusTag?: string;
  dueCount?: number;
}

export function isNativeAndroid() {
  return Capacitor.isNativePlatform() && Capacitor.getPlatform() === 'android';
}

export function getReminderEnabled() {
  return localStorage.getItem(REMINDER_ENABLED_KEY) === '1';
}

export function setReminderEnabled(enabled: boolean) {
  localStorage.setItem(REMINDER_ENABLED_KEY, enabled ? '1' : '0');
}

export function getReminderTime() {
  return localStorage.getItem(REMINDER_TIME_KEY) || '20:00';
}

export function setReminderTime(time: string) {
  localStorage.setItem(REMINDER_TIME_KEY, time);
}

export async function ensureReminderPermissions() {
  const display = await LocalNotifications.checkPermissions();
  if (display.display !== 'granted') {
    const requested = await LocalNotifications.requestPermissions();
    if (requested.display !== 'granted') {
      throw new Error('通知权限未授予');
    }
  }

  if (isNativeAndroid()) {
    const exact = await LocalNotifications.checkExactNotificationSetting();
    if (exact.exact_alarm !== 'granted') {
      await LocalNotifications.changeExactNotificationSetting();
      const next = await LocalNotifications.checkExactNotificationSetting();
      if (next.exact_alarm !== 'granted') {
        throw new Error('请在系统设置中允许精确闹钟，以确保定时提醒准确触发');
      }
    }
  }
}

export async function scheduleDailyReminder(time: string, payload?: ReminderPayload) {
  const [hourText, minuteText] = time.split(':');
  const hour = Number(hourText);
  const minute = Number(minuteText);

  if (!Number.isFinite(hour) || !Number.isFinite(minute) || hour < 0 || hour > 23 || minute < 0 || minute > 59) {
    throw new Error('提醒时间格式无效');
  }

  await LocalNotifications.createChannel({
    id: REMINDER_CHANNEL_ID,
    name: '每日复习提醒',
    description: '每天定时提醒家长安排孩子复习',
    importance: 4,
    visibility: 1,
    vibration: true,
  });

  await LocalNotifications.cancel({
    notifications: [{ id: REMINDER_NOTIFICATION_ID }],
  });

  const focusTag = String(payload?.focusTag || '').trim();
  const dueCount = Math.max(0, Number(payload?.dueCount || 0));
  const body = focusTag
    ? `今天建议优先复习「${focusTag}」标签${dueCount > 0 ? `（到期 ${dueCount} 题）` : ''}。`
    : '今天的错题复习时间到了，和孩子一起开始吧。';

  await LocalNotifications.schedule({
    notifications: [
      {
        id: REMINDER_NOTIFICATION_ID,
        title: '复习提醒',
        body,
        schedule: {
          on: {
            hour,
            minute,
          },
          repeats: true,
          allowWhileIdle: true,
        },
        channelId: REMINDER_CHANNEL_ID,
      },
    ],
  });
}

export async function cancelDailyReminder() {
  await LocalNotifications.cancel({
    notifications: [{ id: REMINDER_NOTIFICATION_ID }],
  });
}
