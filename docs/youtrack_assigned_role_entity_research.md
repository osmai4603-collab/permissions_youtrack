# بحث: كينونة منح دور لمجموعة (AssignedRole) في YouTrack

> **المصدر الرسمي**: YouTrack Developer Portal - REST API Reference - Entities
> **تاريخ البحث**: 2026-10-05
> **آخر تحديث للتوثيق الرسمي**: October 2026
> **متاح منذ**: YouTrack 2026.1

---

## 1. الكينونة الرئيسية: `AssignedRole`

### 1.1 التعريف الرسمي

> "Represents a role assigned to a user or a group in a specific scope."

تمثل **دوراً مُعيّناً** لمستخدم أو مجموعة ضمن نطاق محدد.

### 1.2 الـ Endpoint

```text
POST /api/assignedRoles    — لإنشاء تعيين دور جديد
GET  /api/assignedRoles    — لعرض التعيينات
```

### 1.3 حقول الكينونة (Attributes)

| Field | Type | Description | Read-only? |
| --- | --- | --- | --- |
| `id` | String | معرّف قاعدة البيانات لتعيين الدور | ✅ Read-only |
| `role` | [Role](#2-كينونة-role-الدور) | الدور الذي تم تعيينه | ❌ |
| `scope` | [AccessScope](#3-كينونة-accessscope-نطاق-الوصول) | النطاق الذي يُعيّن فيه الدور (Global, Organization, Project) | ❌ |
| `holder` | [User](#5-كينونة-user-المستخدم) أو [UserGroup](#4-كينونة-usergroup-مجموعة-المستخدمين) | المستخدم أو المجموعة التي تحمل الدور | ✅ Read-only |

### 1.4 مثال على Payload لمنح دور لمجموعة

```json
{
  "role": {
    "id": "role_id_here"
  },
  "scope": {
    "id": "project_or_org_or_global_scope_id",
    "$type": "ProjectScope"
  },
  "holder": {
    "id": "group_id_here",
    "$type": "UserGroup"
  }
}
```

### 1.5 مخطط العلاقات

```text
                    ┌─────────────────┐
                    │  AssignedRole    │
                    │─────────────────│
                    │  id: String     │
                    │  role ──────────┼──────▶ Role
                    │  scope ─────────┼──────▶ AccessScope
                    │  holder ────────┼──────▶ User | UserGroup
                    └─────────────────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
         ┌────────┐   ┌──────────┐   ┌──────────────┐
         │  Role  │   │  Scope   │   │   Holder     │
         │────────│   │──────────│   │──────────────│
         │ id     │   │ Global   │   │ User         │
         │ name   │   │ Org      │   │ UserGroup    │
         │ desc   │   │ Project  │   │  ├ AllUsers   │
         │ perms[]│   └──────────┘   │  ├ Registered │
         │ immut. │                   │  ├ ProjectTeam│
         └────────┘                   │  └ Nested     │
                                      └──────────────┘
```

---

## 2. كينونة `Role` (الدور)

> "Represents a role in YouTrack."

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-Role.html>

| Field | Type | Description | Read-only? |
| --- | --- | --- | --- |
| `id` | String | معرّف قاعدة البيانات للدور | ✅ Read-only |
| `name` | String | اسم الدور | ❌ |
| `description` | String | وصف الدور | ❌ (Can be null) |
| `permissions` | Array of [Permission] | قائمة الصلاحيات المضمّنة في الدور | ✅ Read-only |
| `immutable` | Boolean | `true` لدور System Admin فقط — لا يمكن تعديله أو حذفه | ✅ Read-only |

---

## 3. كينونة `AccessScope` (نطاق الوصول)

> "Represents the scope where a role is assigned in YouTrack. Can be GlobalScope, OrganizationScope, or ProjectScope."

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-AccessScope.html>

### 3.1 الكينونة الأساسية

| Field | Type | Description |
| --- | --- | --- |
| `id` | String | معرّف النطاق (Read-only) |

### 3.2 الأنواع الفرعية (Subtypes)

كينونة `AccessScope` هي كينونة **مجردة (abstract)** تُورّث إلى ثلاثة أنواع:

#### 3.2.1 `GlobalScope`

> "Represents the global scope. Roles assigned in this scope apply across the entire YouTrack."

| Field | Type | Description |
| --- | --- | --- |
| `id` | String | معرّف النطاق (Read-only) |

**ملاحظة**: لا يحتوي حقول إضافية — النطاق العالمي يشمل كل شيء.

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-GlobalScope.html>

#### 3.2.2 `OrganizationScope`

> "Represents an organization scope. Roles assigned in this scope apply within a specific organization."

| Field | Type | Description |
| --- | --- | --- |
| `id` | String | معرّف النطاق (Read-only) |
| `organization` | Organization | المنظمة المرتبطة بهذا النطاق (Can be null) |

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-OrganizationScope.html>

#### 3.2.3 `ProjectScope`

> "Represents a project scope. Roles assigned in this scope apply within a specific project."

| Field | Type | Description |
| --- | --- | --- |
| `id` | String | معرّف النطاق (Read-only) |
| `project` | Project | المشروع المرتبط بهذا النطاق (Can be null) |

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-ProjectScope.html>

---

## 4. كينونة `UserGroup` (مجموعة المستخدمين)

> "Represents a group of users."

**المصدر**: <https://www.jetbrains.com/help/youtrack/devportal/api-entity-UserGroup.html>

### 4.1 الحقول

| Field | Type | Description | Read-only? |
| --- | --- | --- | --- |
| `id` | String | معرّف مجموعة المستخدمين | ✅ Read-only |
| `name` | String | اسم المجموعة | ❌ |
| `ringId` | String | معرّف المجموعة في Hub (للمطابقة بين YouTrack و Hub) | ✅ Read-only (Can be null) |
| `usersCount` | Long | عدد المستخدمين في المجموعة | ✅ Read-only |
| `icon` | String | رابط شعار المجموعة | ✅ Read-only (Can be null) |
| `allUsersGroup` | Boolean | `true` إذا كانت المجموعة تحتوي جميع المستخدمين | ✅ Read-only |
| `users` | Array of User | جميع المستخدمين في المجموعة (بما فيهم المتعدّين من المجموعات الفرعية) | ✅ Read-only |

### 4.2 الأنواع الفرعية (Extended by)

| النوع الفرعي | الوصف |
| --- | --- |
| `AllUsersGroup` | مجموعة جميع المستخدمين |
| `RegisteredUsersGroup` | مجموعة المستخدمين المسجلين |
| `ProjectTeam` | فريق المشروع (متاح منذ YouTrack 2026.1) |
| `NestedGroup` | مجموعة متداخلة |

### 4.3 الموارد المرتبطة (Related Resources)

- `GET /api/groups` — الوصول لمجموعات المستخدمين
- `GET /api/admin/customFieldSettings/bundles/user/{bundleID}/groups` — المجموعات في حزمة المستخدمين

---

## 5. كينونة `User` (المستخدم)

> حامل الدور البديل — يمكن أن يكون holder من نوع User بدلاً من UserGroup.

الحقول الأساسية: `id`, `login`, `fullName`, `email`, `guest`, `online`, `banned`, `tags`, `savedQueries`

---

## 6. ملخص: كيف يعمل منح دور لمجموعة؟

```text
┌──────────────────────────────────────────────────────────┐
│                POST /api/assignedRoles                    │
│──────────────────────────────────────────────────────────│
│                                                          │
│  ┌─── role ────────────────┐                             │
│  │  { "id": "role-xyz" }   │  ◄── أي دور معرّف مسبقاً    │
│  └─────────────────────────┘                             │
│                                                          │
│  ┌─── scope ───────────────────────────────┐             │
│  │  { "id": "scope-id",                   │             │
│  │    "$type": "GlobalScope"        }      │  ◄── أحد:   │
│  │           | "OrganizationScope"         │   Global    │
│  │           | "ProjectScope"              │   Org       │
│  └─────────────────────────────────────────┘   Project   │
│                                                          │
│  ┌─── holder ──────────────────────────────┐             │
│  │  { "id": "group-abc",                  │             │
│  │    "$type": "UserGroup"           }     │  ◄── أحد:   │
│  │           | "User"                      │   User      │
│  └─────────────────────────────────────────┘   Group     │
│                                                          │
│  ═══════════════════════════════════════════             │
│  النتيجة: جميع أعضاء المجموعة يحصلون على                │
│  صلاحيات الدور ضمن النطاق المحدد                        │
└──────────────────────────────────────────────────────────┘
```

---

## 7. القواعد الجوهرية عند التعيين

| القاعدة | التفصيل |
| --- | --- |
| **نطاق عالمي** | جميع صلاحيات الدور تُفعّل |
| **نطاق منظمة** | الصلاحيات ذات النطاق العالمي تُتجاهل |
| **نطاق مشروع** | الصلاحيات العالمية والتنظيمية تُتجاهل. يُمنع التعيين إذا لم يوجد صلاحية واحدة على الأقل بنطاق المشروع |
| **holder = Read-only** | لا يمكن تغيير الحامل بعد الإنشاء — يجب حذف التعيين وإنشاء واحد جديد |
| **Customer Groups** | ❌ لا يمكن تعيين أدوار لمجموعات العملاء مطلقاً |

---

## 8. المصادر الرسمية

| الكينونة | الرابط |
| --- | --- |
| AssignedRole | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-AssignedRole.html> |
| Role | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-Role.html> |
| AccessScope | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-AccessScope.html> |
| GlobalScope | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-GlobalScope.html> |
| OrganizationScope | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-OrganizationScope.html> |
| ProjectScope | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-ProjectScope.html> |
| UserGroup | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-UserGroup.html> |
| Manage Group Access | <https://www.jetbrains.com/help/youtrack/server/configure-access-for-a-user-group.html> |
