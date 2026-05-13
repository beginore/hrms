# HRMS — Workflow Guide

Base URL: http://localhost:8080/v1
Swagger UI: http://localhost:8080/swagger/index.html

Все защищённые эндпоинты требуют заголовок:
  Authorization: Bearer <access_token>

---

## Порядок действий при первом запуске

1. Запустить миграции
2. Зарегистрировать организацию
3. Создать отделы и должности
4. Пригласить сотрудников
5. Настроить графики работы
6. Работать с посещаемостью

---

## 1. ОРГАНИЗАЦИЯ — регистрация

### Шаг 1. Зарегистрировать организацию

POST /v1/organizations
(публичный, токен не нужен)

```json
{
  "name": "Acme Corp",
  "vat_id": "123456789",
  "description": "IT компания",
  "address": "ул. Абая 1",
  "city_id": "1",
  "admin_email": "admin@acme.kz",
  "admin_first_name": "Иван",
  "admin_last_name": "Иванов",
  "admin_phone": "+77001234567",
  "password": "SecurePass123!"
}
```

Ответ: organization_id, otp отправляется на email

### Шаг 2. Подтвердить email через OTP

POST /v1/organizations/verify-otp

```json
{
  "email": "admin@acme.kz",
  "otp": "123456"
}
```

### Шаг 3. Войти

POST /v1/auth/login

```json
{
  "email": "admin@acme.kz",
  "password": "SecurePass123!"
}
```

Ответ: access_token, refresh_token, id_token

Сохрани access_token — он используется во всех защищённых запросах.

---

## 2. СПРАВОЧНИКИ — отделы и должности

Нужны до создания сотрудников.

### Создать отдел

POST /v1/organizations/departments

```json
{ "name": "Разработка" }
```

Ответ: { "id": "uuid" }

### Список отделов

GET /v1/organizations/departments

### Удалить отдел

DELETE /v1/organizations/departments/:id

### Создать должность

POST /v1/organizations/positions

```json
{ "name": "Backend Developer" }
```

### Список должностей

GET /v1/organizations/positions

### Удалить должность

DELETE /v1/organizations/positions/:id

---

## 3. INVITE — приглашение сотрудников

Сотрудники НЕ регистрируются напрямую. HR генерирует приглашение, сотрудник переходит по ссылке.

### Шаг 1. Сгенерировать приглашение (публичный)

POST /v1/invites/generate

```json
{
  "org_id": "uuid-организации",
  "email": "employee@acme.kz",
  "first_name": "Пётр",
  "last_name": "Петров",
  "role": "Employee",
  "position": "Backend Developer"
}
```

Ответ: invite_code (также отправляется на email)

### Шаг 2. Проверить инвайт (публичный)

POST /v1/invites/verify

```json
{ "code": "INVITE-CODE-HERE" }
```

### Шаг 3. Завершить регистрацию (публичный)

POST /v1/invites/complete-registration

```json
{
  "code": "INVITE-CODE-HERE",
  "password": "MyPassword123!",
  "phone_number": "+77009876543"
}
```

После этого сотрудник может войти через POST /v1/auth/login.

---

## 4. EMPLOYEES — управление сотрудниками

После того как сотрудник принял инвайт и залогинился, HR создаёт его запись в системе.

### Создать сотрудника

POST /v1/employees

```json
{
  "user_id": "uuid-пользователя",
  "department_id": "uuid-отдела",
  "position_id": "uuid-должности",
  "role": "Employee",
  "salary_rate": "500000",
  "status": "Active"
}
```

Поле role: Employee | Manager | HR | Admin

### Список сотрудников организации

GET /v1/employees

### Получить сотрудника

GET /v1/employees/:id

### Изменить роль

PATCH /v1/employees/:id/role
```json
{ "role": "Manager" }
```

### Изменить зарплату

PATCH /v1/employees/:id/salary
```json
{ "salary_rate": "600000" }
```

### Изменить статус

PATCH /v1/employees/:id/status
```json
{ "status": "Inactive" }
```

### Перевести в другой отдел

PATCH /v1/employees/:id/department
```json
{ "department_id": "uuid-нового-отдела" }
```

### Изменить должность

PATCH /v1/employees/:id/position
```json
{ "position_id": "uuid-новой-должности" }
```

### Удалить сотрудника

DELETE /v1/employees/:id

---

## 5. AUTH — токены и сессии

### Войти

POST /v1/auth/login
```json
{
  "email": "user@acme.kz",
  "password": "Password123!"
}
```

### Обновить токен

POST /v1/auth/refresh
```json
{ "refresh_token": "..." }
```

### Выйти (защищённый)

POST /v1/auth/logout
```json
{ "refresh_token": "..." }
```

### Сбросить пароль — запросить код

POST /v1/auth/forgot-password
```json
{ "email": "user@acme.kz" }
```

### Сбросить пароль — подтвердить

POST /v1/auth/reset-password
```json
{
  "email": "user@acme.kz",
  "code": "123456",
  "new_password": "NewPassword123!"
}
```

---

## 6. ATTENDANCE — посещаемость

### 6.1 Настроить гибкий график сотруднику

Делается один раз при найме. Если не настроено — используется дефолт 09:00–18:00 с порогом 15 мин.

POST /v1/attendance/work-schedules

```json
{
  "employee_id": "uuid-сотрудника",
  "work_start": "09:00",
  "work_end": "18:00",
  "late_threshold_minutes": 15
}
```

Повторный вызов — обновляет график (upsert).

### Получить график сотрудника

GET /v1/attendance/work-schedules/:employee_id

---

### 6.2 СКУД — фиксация входа/выхода

В реальном проекте СКУД сам дёргает этот эндпоинт. Сейчас можно симулировать вручную.

POST /v1/attendance/skud-events

Вход сотрудника:
```json
{
  "employee_id": "uuid-сотрудника",
  "event_type": "ENTER",
  "device_id": "turnstile-01",
  "occurred_at": "2026-05-12T09:05:00+06:00"
}
```

Выход сотрудника:
```json
{
  "employee_id": "uuid-сотрудника",
  "event_type": "EXIT",
  "device_id": "turnstile-01",
  "occurred_at": "2026-05-12T18:10:00+06:00"
}
```

Поле occurred_at — опциональное. Если не передать, берётся текущее время.

Что происходит автоматически:
- Создаётся attendance_record для этого дня
- Если пришёл позже work_start + late_threshold_minutes → status = LATE
- Если на этот день есть одобренная заявка (отпуск/больничный) → тип берётся из заявки, СКУД игнорируется
- Повторный ENTER/EXIT в тот же день — обновляет запись (не дублирует)

---

### 6.3 Заявки на отсутствие

Типы заявок:
- SICK_LEAVE      — больничный
- VACATION        — отпуск
- REMOTE          — удалённая работа
- BUSINESS_TRIP   — командировка
- UNPAID_LEAVE    — отгул за свой счёт

#### Подать заявку (сотрудник)

POST /v1/attendance/leave-requests

```json
{
  "employee_id": "uuid-сотрудника",
  "type": "SICK_LEAVE",
  "start_date": "2026-05-13",
  "end_date": "2026-05-15",
  "reason": "ОРВИ",
  "document_url": "https://storage.example.com/sick-leave-scan.pdf"
}
```

Для REMOTE/VACATION document_url необязателен.

#### Список всех заявок организации (HR/Manager)

GET /v1/attendance/leave-requests

#### Получить конкретную заявку

GET /v1/attendance/leave-requests/:id

#### Одобрить или отклонить (HR или руководитель)

PATCH /v1/attendance/leave-requests/:id/review

```json
{ "action": "approve" }
```

или

```json
{ "action": "reject" }
```

Статус заявки:
- PENDING   — подана, ожидает решения
- APPROVED  — одобрена
- REJECTED  — отклонена

После одобрения: если СКУД получит событие на эти даты — он увидит заявку и проставит attendance_record с правильным типом.

---

### 6.4 Просмотр посещаемости

#### Все сотрудники организации за период

GET /v1/attendance?start_date=2026-05-01&end_date=2026-05-31

Если параметры не переданы — возвращает текущий месяц.

#### Посещаемость конкретного сотрудника

GET /v1/attendance/employees/:employee_id?start_date=2026-05-01&end_date=2026-05-31

Поля в ответе:
```json
{
  "id": "uuid",
  "employee_id": "uuid",
  "date": "2026-05-12",
  "type": "OFFICE",
  "source": "SKUD",
  "status": "LATE",
  "check_in": "2026-05-12T09:20:00+06:00",
  "check_out": "2026-05-12T18:05:00+06:00"
}
```

Значения type:   OFFICE | REMOTE | SICK_LEAVE | VACATION | BUSINESS_TRIP | ABSENT
Значения source: SKUD | MANUAL | SYSTEM
Значения status: PRESENT | LATE | ABSENT | ON_LEAVE

---

## Типичный рабочий день — полный цикл

1. Сотрудник приходит → СКУД шлёт ENTER → запись создана, status=PRESENT или LATE
2. Сотрудник уходит → СКУД шлёт EXIT → check_out проставлен
3. Заболел → подаёт SICK_LEAVE заявку → HR одобряет
4. На дни больничного СКУД (если придёт) не перезапишет — тип ON_LEAVE сохранится
5. HR смотрит отчёт GET /v1/attendance за месяц

---

## Справочник — все эндпоинты

### Публичные (без токена)
```
POST   /v1/auth/login
POST   /v1/auth/refresh
POST   /v1/auth/forgot-password
POST   /v1/auth/reset-password
POST   /v1/organizations
POST   /v1/organizations/verify-otp
POST   /v1/invites/generate
POST   /v1/invites/verify
POST   /v1/invites/complete-registration
GET    /v1/legal/documents
GET    /v1/cities
```

### Защищённые (нужен Bearer токен)
```
POST   /v1/auth/logout

POST   /v1/organizations/consents
GET    /v1/organizations/consents/validate

POST   /v1/organizations/departments
GET    /v1/organizations/departments
DELETE /v1/organizations/departments/:id

POST   /v1/organizations/positions
GET    /v1/organizations/positions
DELETE /v1/organizations/positions/:id

POST   /v1/employees
GET    /v1/employees
GET    /v1/employees/:id
PATCH  /v1/employees/:id/role
PATCH  /v1/employees/:id/salary
PATCH  /v1/employees/:id/status
PATCH  /v1/employees/:id/department
PATCH  /v1/employees/:id/position
DELETE /v1/employees/:id

POST   /v1/attendance/work-schedules
GET    /v1/attendance/work-schedules/:employee_id

POST   /v1/attendance/skud-events

POST   /v1/attendance/leave-requests
GET    /v1/attendance/leave-requests
GET    /v1/attendance/leave-requests/:id
PATCH  /v1/attendance/leave-requests/:id/review

GET    /v1/attendance
GET    /v1/attendance/employees/:employee_id
```
