# دليل شامل عن نظام الأدوار (Roles) في YouTrack

> **المصدر الرسمي:** [YouTrack Cloud 2026.2 Documentation](https://www.jetbrains.com/help/youtrack/cloud/)
> **تاريخ آخر تحديث للمصدر:** 23 سبتمبر 2026
> **تاريخ إعداد هذا الدليل:** 3 أكتوبر 2026

---

## جدول المحتويات

1. [مفهوم الدور (Role) في YouTrack](#1-مفهوم-الدور-role-في-youtrack)
2. [دورة حياة الدور (Role Lifecycle)](#2-دورة-حياة-الدور-role-lifecycle)
3. [نطاقات الصلاحيات (Permission Scopes)](#3-نطاقات-الصلاحيات-permission-scopes)
4. [الأدوار الافتراضية (Default Roles)](#4-الأدوار-الافتراضية-default-roles)
5. [جدول مقارنة الصلاحيات لجميع الأدوار](#5-جدول-مقارنة-الصلاحيات-لجميع-الأدوار)
6. [إنشاء وتعديل الأدوار](#6-إنشاء-وتعديل-الأدوار)
7. [نسخ الأدوار (Clone Roles)](#7-نسخ-الأدوار-clone-roles)
8. [دمج الأدوار (Merge Roles)](#8-دمج-الأدوار-merge-roles)
9. [إسناد الأدوار للمستخدمين والمجموعات](#9-إسناد-الأدوار-للمستخدمين-والمجموعات)
10. [الحماية من تصعيد الصلاحيات](#10-الحماية-من-تصعيد-الصلاحيات)
11. [ملاحظات الهجرة والتحديث](#11-ملاحظات-الهجرة-والتحديث)
12. [المصادر الرسمية](#12-المصادر-الرسمية)

---

## 1. مفهوم الدور (Role) في YouTrack

### التعريف الرسمي

> **Role** (الدور): حاوية مُسمّاة لمجموعة من الصلاحيات (Permissions) يتم إسنادها للمستخدمين أو المجموعات ضمن نطاق مشروع أو منظمة.
>
> *"A role in YouTrack is a container for a set of permissions. Roles are assigned to users and groups on a per-project basis."*

### الخصائص الجوهرية للدور

| الخاصية | الوصف |
| --------- | ------- |
| **الاسم (Name)** | اسم فريد يُعرّف الدور في النظام |
| **الوصف (Description)** | نص اختياري يشرح الغرض من الدور |
| **الصلاحيات (Permissions)** | مجموعة الصلاحيات المضمّنة في الدور |
| **النطاق (Scope)** | المستوى الذي يُفعّل فيه الدور (عام / منظمة / مشروع) |
| **حالة القراءة فقط (Read-only)** | الأدوار الافتراضية المدمجة للقراءة فقط ولا يمكن تعديلها مباشرة |

### العلاقة بين الدور والصلاحية

```text
┌─────────────────────────────────────────────────┐
│                   Role (الدور)                   │
│  ┌─────────────┐ ┌─────────────┐ ┌───────────┐  │
│  │ Permission 1│ │ Permission 2│ │Permission N│  │
│  │  (صلاحية)   │ │  (صلاحية)   │ │ (صلاحية)  │  │
│  └─────────────┘ └─────────────┘ └───────────┘  │
└─────────────────────────────────────────────────┘
         │                    │
         ▼                    ▼
   ┌──────────┐        ┌──────────────┐
   │ User     │        │ Group        │
   │(مستخدم)  │        │ (مجموعة)     │
   └──────────┘        └──────────────┘
         │                    │
         ▼                    ▼
   ┌──────────────────────────────────┐
   │     Project / Organization      │
   │     (مشروع / منظمة)             │
   └──────────────────────────────────┘
```

- **الدور** هو الوحدة التنظيمية الأساسية لإدارة الوصول.
- **الصلاحيات** لا تُمنح مباشرة للمستخدمين، بل تُضاف داخل أدوار ثم تُسند الأدوار.
- يمكن للمستخدم أو المجموعة أن يحمل **أدوار متعددة** في مشاريع مختلفة.
- يمكن للمستخدمين والمجموعات أن **يرثوا الأدوار** عبر عضويتهم في فريق المشروع (Project Team).

---

## 2. دورة حياة الدور (Role Lifecycle)

```text
    ┌─────────────────┐
    │  🆕 إنشاء الدور │
    │  Create Role     │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────────┐
    │ ⚙️ تهيئة الصلاحيات  │
    │ Configure Permissions│
    └────────┬────────────┘
             │
             ▼
    ┌──────────────────────────┐
    │ 📋 إسناد الدور           │
    │ Assign to Users/Groups   │
    └────────┬─────────────────┘
             │
             ▼
    ┌──────────────────────┐
    │ 🔄 مرحلة التشغيل    │◄──────────────┐
    │ Active Usage          │               │
    └────────┬─────────────┘               │
             │                              │
             ▼                              │
    ┌──────────────────┐                   │
    │ هل يحتاج تعديل؟ │──── نعم ─────┐   │
    └────────┬─────────┘              │   │
             │ لا                      │   │
             │           ┌─────────────┤   │
             │           │  ✏️ تعديل   │───┘
             │           │  ─────────  │
             │           │  📑 نسخ    │
             │           │  ─────────  │
             │           │  🔀 دمج    │
             │           └─────────────┘
             ▼
    ┌──────────────────┐
    │ 🗑️ حذف الدور    │
    │ Delete Role       │
    └──────────────────┘
```

### المراحل التفصيلية

#### المرحلة 1: الإنشاء (Creation)

- يمكن إنشاء دور جديد من صفحة الإدارة: `Administration > Access Management > Roles`
- يتطلب صلاحية: **Low-level Admin Write**
- يتم تحديد الاسم والوصف والصلاحيات عند الإنشاء
- بديل: **نسخ (Clone)** دور موجود لإنشاء دور مشابه بتعديلات طفيفة

#### المرحلة 2: التهيئة (Configuration)

- إضافة أو إزالة الصلاحيات من الدور
- يمكن تجميع الصلاحيات حسب **الكيان (Entity)** أو **العملية (Operation)**
- التغييرات تُحفظ تلقائياً عند التعديل

#### المرحلة 3: الإسناد (Assignment)

- إسناد الدور لمستخدمين أو مجموعات على مستوى مشروع أو منظمة أو عالمياً
- يمكن الإسناد عبر فريق المشروع (Project Team)
- الدور يجب أن يحتوي **صلاحية واحدة على الأقل بنطاق مشروع** ليتم إسناده على مستوى المشروع

#### المرحلة 4: التشغيل (Active Usage)

- المستخدمون يمارسون صلاحياتهم المكتسبة من الأدوار
- النظام يتحقق من الصلاحيات عند كل عملية

#### المرحلة 5: التعديل / الصيانة (Modification)

- **تعديل مباشر**: إضافة أو إزالة صلاحيات من دور مخصص (الأدوار الافتراضية للقراءة فقط)
- **نسخ (Clone)**: إنشاء نسخة قابلة للتعديل من دور (بما في ذلك الأدوار الافتراضية)
- **دمج (Merge)**: دمج دورين أو أكثر في دور واحد (يتم نقل جميع الإسنادات)

#### المرحلة 6: الإزالة (Deletion)

- حذف الأدوار غير المستخدمة
- **لا يمكن حذف الأدوار الافتراضية المدمجة**

---

## 3. نطاقات الصلاحيات (Permission Scopes)

يفصل YouTrack بين **تعريف الدور** و**إسناد الدور**. يمكن إنشاء أدوار بنطاقات مختلطة، لكن النظام يُفعّل فقط الصلاحيات المتوافقة مع مستوى الإسناد.

### مستويات النطاق

| النطاق | الوصف | مثال |
| -------- | ------- | ------ |
| **Global (عالمي)** | ينطبق على النظام بأكمله | Create User, Create Project, Low-level Admin |
| **Organization (منظمة)** | محدود بمنظمة معينة | Read Organization, Update Organization |
| **Project (مشروع)** | محدود بمشروع معين | Read Issue, Update Issue, Create Article |

### قواعد تفعيل الصلاحيات حسب مستوى الإسناد

| مستوى الإسناد | الصلاحيات المُفعّلة |
| --------------- | --------------------- |
| **عالمي (Global)** | جميع الصلاحيات (عالمية + منظمة + مشروع) تكون صالحة |
| **منظمة (Organization)** | صلاحيات المنظمة والمشروع فقط. الصلاحيات العالمية **تُتجاهل** |
| **مشروع (Project)** | صلاحيات المشروع فقط. الصلاحيات العالمية والتنظيمية **تُتجاهل**. يُمنع الإسناد ما لم يحتوي الدور على صلاحية مشروع واحدة على الأقل |

> **ملاحظة هامة:** هذا الفصل بين التعريف والإسناد يمنح مرونة عالية - يمكن إنشاء دور واحد بصلاحيات مختلطة واستخدامه في مستويات مختلفة، حيث يُفعّل النظام تلقائياً الصلاحيات المناسبة للمستوى.

---

## 4. الأدوار الافتراضية (Default Roles)

يوفر YouTrack **6 أدوار افتراضية مُدمجة** (Predefined) وهي **للقراءة فقط** - لتخصيصها يجب نسخها أولاً (Clone).

### نظرة عامة سريعة

| # | الدور | الغرض | نطاق الصلاحيات | عدد الصلاحيات |
| --- | ------- | ------- | ---------------- | --------------- |
| 1 | **System Admin** | إدارة النظام بالكامل | جميع النطاقات | جميع الصلاحيات |
| 2 | **Project Admin** | إدارة المشاريع وإعداداتها | مشروع | ~30 صلاحية |
| 3 | **Contributor** | العمل اليومي على المهام والمقالات | مشروع | ~22 صلاحية |
| 4 | **Observer** | مراقبة أساسية وقراءة بيانات المستخدمين | عالمي | 3 صلاحيات |
| 5 | **User Manager** | إنشاء حسابات مستخدمين | عالمي | 1 صلاحية |
| 6 | **Project Creator** | إنشاء مشاريع جديدة | عالمي | 1 صلاحية |

---

### 4.1 System Admin (مدير النظام)

**الغرض:** مخصص للمستخدمين المسؤولين عن إدارة بيئة YouTrack بالكامل.

**الخصائص:**

- يمتلك **جميع الصلاحيات المتاحة** في YouTrack عبر جميع النطاقات
- هو الدور الوحيد الذي يملك صلاحيات مثل: `Override Visibility Restrictions`, `Delete Project`, `Low-level Admin Read/Write`
- يملك صلاحيات المشروع على المستوى العالمي، مما يعني أنه يملك صلاحية تنفيذ الإجراءات المتعلقة بالمشاريع في **جميع المشاريع**

**الصلاحيات الحصرية (متوفرة فقط في System Admin افتراضياً):**

| الكيان | الصلاحية |
| -------- | ---------- |
| Application | Low-level Admin Read |
| Application | Low-level Admin Write |
| User | Create User |
| User | Delete User |
| User | Update User |
| Organization | Create Organization |
| Organization | Read Organization |
| Organization | Update Organization |
| Organization | Delete Organization |
| Project | Create Project |
| Project | Delete Project |
| Visibility | Override Visibility Restrictions |

---

### 4.2 Project Admin (مدير المشروع)

**الغرض:** للمستخدمين الذين يُنشئون ويُديرون المشاريع. يملك نفس صلاحيات Contributor بالإضافة إلى إدارة إعدادات المشروع.

**الخصائص:**

- جميع صلاحياته **بنطاق مشروع** (Project-scoped)
- لا يتضمن صلاحيات عالمية مثل `Create Project` أو `Create User`
- لأنه يتضمن صلاحية `Update Project`، يمكنه إدارة: workflows, project-level apps, integrations, agile boards
- عند تفعيل ميزة `Per-project Notification Template Customization` يمكنه تعديل قوالب الإشعارات

**صلاحيات دور Project Admin:**

| الكيان | الصلاحيات |
| -------- | ----------- |
| **Project** | Read Project Basic, Read Project Full, Update Project |
| **Issue** | Read Issue, Read Issue Private Fields, Update Issue, Create Issue, Delete Issue, Link Issues, Update Issue Private Fields, Apply Commands Silently, View Watchers, Update Watchers, View Voters |
| **Attachment** | Add Attachment, Update Attachment, Delete Attachment |
| **Comment** | Create Issue Comment, Read Issue Comment, Update Issue Comment, Delete Issue Comment, Update Not Own Issue Comment, Delete Not Own and Permanent Comment Delete, Read Article Comment, Create Article Comment, Update Article Comment, Delete Article Comment |
| **Work Item** | Read Work Item, Update Work Item, Update Not Own Work Item, Create Work Item, Create Not Own Work Item |
| **Article** | Read Article, Create Article, Update Article, Delete Article |

---

### 4.3 Contributor (المُساهم)

**الغرض:** للمستخدمين الذين يُنشئون ويعملون على المهام. يمكنهم إنشاء مهام جديدة وإضافة تعليقات وعرض المهام الموجودة وتحديث معظم سمات المهام.

**الخصائص:**

- هو **الدور الافتراضي** لأعضاء فريق المشاريع الجديدة (بدءاً من الإصدار 2023.1)
- يحل محل دور `Developer` القديم (الذي كان يفتقر لصلاحية `Read User Details`)
- جميع صلاحياته **بنطاق مشروع**

> **ملاحظة تاريخية:** بدءاً من الإصدار 2023.1، حل `Contributor` محل `Developer`. في الأنظمة المُحدّثة من إصدارات أقدم، يتواجد الدوران معاً لمنع تصعيد الصلاحيات. يبقى `Developer` في المشاريع الموجودة، بينما يُستخدم `Contributor` للمشاريع الجديدة.

**صلاحيات دور Contributor:**

| الكيان | الصلاحيات |
| -------- | ----------- |
| **Project** | Read Project Basic |
| **Issue** | Read Issue, Read Issue Private Fields, Update Issue, Create Issue, Delete Issue, Link Issues, Update Issue Private Fields, View Watchers, Update Watchers, View Voters |
| **Attachment** | Add Attachment, Update Attachment, Delete Attachment |
| **Comment** | Create Issue Comment, Read Issue Comment, Update Issue Comment, Delete Issue Comment, Read Article Comment, Create Article Comment |
| **Work Item** | Read Work Item, Update Work Item, Create Work Item |
| **Article** | Read Article, Create Article |

---

### 4.4 Observer (المُراقب)

**الغرض:** يمنح وصولاً أساسياً للتفاعل مع المستخدمين المسجلين الآخرين في النظام. يمكن للمراقبين عرض بيانات الملف الشخصي وتحديث ملفاتهم الشخصية.

**الخصائص:**

- صلاحياته **بنطاق عالمي** (Global-scoped)
- يُعتبر الدور الأساسي الذي يُمنح لجميع المستخدمين المسجلين عادةً
- لا يمنح أي وصول لمحتوى المشاريع

**صلاحيات دور Observer:**

| الكيان | الصلاحيات |
|--------|-----------|
| **User Profile** | Update Self |
| **User** | Read User Basic, Read User Details |

---

### 4.5 User Manager (مدير المستخدمين)

**الغرض:** يتيح للمستخدمين إنشاء حسابات لمستخدمين آخرين ودعوتهم للانضمام إلى YouTrack.

**الخصائص:**

- يملك صلاحية واحدة فقط بنطاق عالمي
- يمنح فقط القدرة على **إنشاء** حسابات جديدة
- **لا يمنح** صلاحية قراءة أو تعديل الحسابات الأخرى (هذه تأتي من أدوار أخرى)
- صلاحية قراءة حسابات المستخدمين تُمنح عبر دور **Observer**
- صلاحية تعديل حسابات المستخدمين الآخرين محصورة بـ **System Admin**

**صلاحيات دور User Manager:**

| الكيان | الصلاحيات |
|--------|-----------|
| **User** | Create User |

---

### 4.6 Project Creator (منشئ المشاريع)

**الغرض:** يمنح المستخدمين القدرة على إنشاء مشاريع جديدة في النظام دون الوصول لمحتوى مشاريع الآخرين.

**الخصائص:**

- يملك صلاحية واحدة فقط بنطاق عالمي
- مالكو المشاريع يحصلون تلقائياً على دور **Project Admin** في مشاريعهم الخاصة
- مفيد لمنح غير المديرين القدرة على إنشاء مشاريعهم الخاصة

**صلاحيات دور Project Creator:**

| الكيان | الصلاحيات |
|--------|-----------|
| **Project** | Create Project |

---

## 5. جدول مقارنة الصلاحيات لجميع الأدوار

### 5.1 صلاحيات النطاق العالمي (Global Scope)

| الكيان | الصلاحية | System Admin | Observer | Project Creator | User Manager |
| -------- | ---------- | :-----------: | :--------: | :--------------: | :------------: |
| Application | Low-level Admin Read | ✅ | — | — | — |
| Application | Low-level Admin Write | ✅ | — | — | — |
| Project | Create Project | ✅ | — | ✅ | — |
| User | Create User | ✅ | — | — | ✅ |
| User | Delete User | ✅ | — | — | — |
| User | Read User Details | ✅ | ✅ | — | — |
| User | Read User Basic | ✅ | ✅ | — | — |
| User | Update Self | ✅ | ✅ | — | — |
| User | Update User | ✅ | — | — | — |
| Organization | Create Organization | ✅ | — | — | — |

### 5.2 صلاحيات النطاق التنظيمي (Organization Scope)

> بشكل افتراضي، صلاحيات نطاق المنظمة مُسندة **فقط** لدور System Admin.

| الصلاحية | System Admin | الأدوار الأخرى |
| ---------- | :-----------: | :--------------: |
| Read Organization | ✅ | — |
| Update Organization | ✅ | — |
| Delete Organization | ✅ | — |

### 5.3 صلاحيات نطاق المشروع (Project Scope)

| الكيان | الصلاحية | System Admin | Project Admin | Contributor |
| -------- | ---------- | :-----------: | :------------: | :-----------: |
| **Project** | Read Project Basic | ✅ | ✅ | ✅ |
| | Read Project Full | ✅ | ✅ | — |
| | Update Project | ✅ | ✅ | — |
| | Delete Project | ✅ | — | — |
| **Issue** | Read Issue | ✅ | ✅ | ✅ |
| | Read Issue Private Fields | ✅ | ✅ | ✅ |
| | Update Issue | ✅ | ✅ | ✅ |
| | Create Issue | ✅ | ✅ | ✅ |
| | Delete Issue | ✅ | ✅ | ✅ |
| | Link Issues | ✅ | ✅ | ✅ |
| | Update Issue Private Fields | ✅ | ✅ | ✅ |
| | Apply Commands Silently | ✅ | ✅ | — |
| | View Watchers | ✅ | ✅ | ✅ |
| | Update Watchers | ✅ | ✅ | ✅ |
| | View Voters | ✅ | ✅ | ✅ |
| **Attachment** | Add Attachment | ✅ | ✅ | ✅ |
| | Update Attachment | ✅ | ✅ | ✅ |
| | Delete Attachment | ✅ | ✅ | ✅ |
| **Comment** | Create Issue Comment | ✅ | ✅ | ✅ |
| | Read Issue Comment | ✅ | ✅ | ✅ |
| | Update Issue Comment | ✅ | ✅ | ✅ |
| | Delete Issue Comment | ✅ | ✅ | ✅ |
| | Update Not Own Issue Comment | ✅ | ✅ | — |
| | Delete Not Own and Permanent Comment Delete | ✅ | ✅ | — |
| | Create Article Comment | ✅ | ✅ | ✅ |
| | Read Article Comment | ✅ | ✅ | ✅ |
| | Update Article Comment | ✅ | ✅ | — |
| | Delete Article Comment | ✅ | ✅ | — |
| **Visibility** | Override Visibility Restrictions | ✅ | — | — |
| **Work Item** | Read Work Item | ✅ | ✅ | ✅ |
| | Update Work Item | ✅ | ✅ | ✅ |
| | Create Work Item | ✅ | ✅ | ✅ |
| | Update Not Own Work Item | ✅ | ✅ | — |
| | Create Not Own Work Item | ✅ | ✅ | — |
| **Article** | Read Article | ✅ | ✅ | ✅ |
| | Create Article | ✅ | ✅ | ✅ |
| | Update Article | ✅ | ✅ | — |
| | Delete Article | ✅ | ✅ | — |

### 5.4 الفروقات الرئيسية بين Project Admin و Contributor

| الصلاحية | Project Admin | Contributor | الملاحظة |
| ---------- | :------------: | :-----------: | ---------- |
| Read Project Full | ✅ | — | الوصول الكامل لإعدادات المشروع |
| Update Project | ✅ | — | تعديل إعدادات المشروع |
| Apply Commands Silently | ✅ | — | تنفيذ أوامر بدون إشعارات |
| Update Not Own Issue Comment | ✅ | — | تعديل تعليقات الآخرين |
| Delete Not Own and Permanent Comment Delete | ✅ | — | حذف تعليقات الآخرين نهائياً |
| Update Not Own Work Item | ✅ | — | تعديل سجلات عمل الآخرين |
| Create Not Own Work Item | ✅ | — | إنشاء سجلات عمل باسم الآخرين |
| Update Article | ✅ | — | تعديل المقالات |
| Delete Article | ✅ | — | حذف المقالات |
| Update Article Comment | ✅ | — | تعديل تعليقات المقالات |
| Delete Article Comment | ✅ | — | حذف تعليقات المقالات |

---

## 6. إنشاء وتعديل الأدوار

### 6.1 إنشاء دور جديد (Create a Role)

**المتطلبات:** صلاحية `Low-level Admin Write`

**الخطوات:**

1. من القائمة الرئيسية: `⚙ Administration > Access Management > Roles`
2. انقر على زر **New role**
3. أدخل اسم الدور الجديد
4. أدخل وصفاً اختيارياً
5. اختر الصلاحيات المطلوبة من قسم **Permissions**:
   - استخدم مربع البحث لتصفية الصلاحيات بالاسم
   - استخدم خيار **Group by** لتجميع الصلاحيات حسب `Entity` أو `Operation`
6. انقر على زر **Create**

### 6.2 تعديل دور موجود (Edit an Existing Role)

**المتطلبات:** صلاحية `Low-level Admin Write`

> ⚠️ **تنبيه:** الأدوار الافتراضية المُدمجة **للقراءة فقط**. لتعديلها يجب نسخها أولاً (Clone).

**الخطوات:**

1. من القائمة الرئيسية: `⚙ Administration > Access Management > Roles`
2. اختر الدور من القائمة
3. لتعديل الاسم أو الوصف: انقر على أيقونة القلم ✏️ ثم عدّل واحفظ
4. لتعديل الصلاحيات: فعّل أو عطّل الصلاحيات في قسم **Permissions**
   - التغييرات **تُحفظ تلقائياً**

---

## 7. نسخ الأدوار (Clone Roles)

### متى تستخدم النسخ؟

- عندما تريد تخصيص دور افتراضي (لأنه للقراءة فقط)
- عندما تريد إنشاء دور مشابه لدور موجود مع تعديلات طفيفة
- لإنشاء نقطة بداية سريعة بدلاً من البدء من الصفر

### كيف يعمل النسخ؟

- يُنشئ دوراً جديداً بنفس مجموعة الصلاحيات للدور الأصلي
- الدور المنسوخ **قابل للتعديل بالكامل**
- النسخة **مستقلة تماماً** - تعديلها لا يؤثر على الأصل
- الاسم الافتراضي للنسخة يتضمن "(copy)" أو رقم تسلسلي

### الخطوات

1. من القائمة الرئيسية: `⚙ Administration > Access Management > Roles`
2. اختر الدور المراد نسخه
3. انقر على إجراء **Clone**
4. عدّل الاسم والوصف والصلاحيات حسب الحاجة

---

## 8. دمج الأدوار (Merge Roles)

### متى تستخدم الدمج؟

- عندما يكون لديك أدوار متعددة بصلاحيات متشابهة
- لتبسيط هيكل الأدوار في النظام
- لتوحيد أدوار متشابهة بعد فترة نمو عضوي

### كيف يعمل الدمج؟

- يتم اختيار دور **مستهدف** (Target) ودور أو أكثر **للدمج**
- جميع إسنادات الأدوار المدمجة (المستخدمين والمجموعات في المشاريع) تُنقل إلى الدور المستهدف
- صلاحيات الدور المستهدف **لا تتغير** - يتم فقط نقل الإسنادات
- الأدوار المدمجة **تُحذف** بعد الدمج

> ⚠️ **تحذير:** عملية الدمج **لا يمكن التراجع عنها**. تأكد من مراجعة الفروقات قبل الدمج.

---

## 9. إسناد الأدوار للمستخدمين والمجموعات

### طرق إسناد الأدوار

| الطريقة | الوصف |
| --------- | ------- |
| **إسناد مباشر للمستخدم** | إسناد دور لمستخدم محدد في مشروع أو منظمة معينة |
| **إسناد لمجموعة** | إسناد دور لمجموعة - جميع أعضاء المجموعة يرثون الصلاحيات |
| **فريق المشروع** | إضافة مستخدم أو مجموعة لفريق المشروع مع دور محدد |
| **الإسناد العالمي** | إسناد دور على المستوى العالمي - ينطبق على جميع المشاريع/المنظمات |

### مسار الإسناد عبر فريق المشروع

```
Project Team (فريق المشروع)
    └── Member/Group (عضو/مجموعة)
            └── Role (الدور المُسند)
                    └── Permissions (الصلاحيات المُفعّلة في هذا المشروع فقط)
```

### إدارة الإسنادات عبر المجموعات

يمكن إدارة أدوار المجموعات من صفحة المجموعة في تبويب **Roles**:

- عرض الأدوار المُسندة حالياً ونطاقها
- المشاريع أو المنظمات التي تنطبق فيها
- أعضاء المجموعة يرثون الصلاحيات تلقائياً

---

## 10. الحماية من تصعيد الصلاحيات

يتضمن YouTrack آلية حماية مدمجة ضد **تصعيد الصلاحيات (Permission Escalation)**:

### القواعد الأمنية

1. **لا يمكن منح ما لا تملك:** المستخدمون الذين يملكون صلاحية تعديل الأدوار يمكنهم **إزالة** أي صلاحية من دور، لكنهم **لا يمكنهم إضافة** صلاحيات غير متاحة لحساباتهم الخاصة.

2. **الأدوار الافتراضية محمية:** لا يمكن تعديلها مباشرة، مما يمنع التغييرات العرضية التي قد تؤثر على جميع المستخدمين.

3. **التحقق عند الإسناد على مستوى المشروع:** لا يمكن إسناد دور على مستوى المشروع ما لم يحتوي على صلاحية واحدة على الأقل بنطاق مشروع.

4. **التعايش بين Developer و Contributor:** في الأنظمة المُحدّثة، يتعايش الدوران لمنع تصعيد غير مقصود (Developer يبقى في المشاريع القديمة، Contributor للمشاريع الجديدة).

---

## 11. ملاحظات الهجرة والتحديث

### التحديثات الرئيسية في نظام الأدوار

| الإصدار | التغيير |
| --------- | --------- |
| **2023.1** | استبدال دور `Developer` بـ `Contributor` (مع إضافة `Read User Details`). الدوران يتعايشان في الأنظمة المُحدّثة |
| **2026.1** | تبسيط نموذج الصلاحيات لتقليل التعقيد |
| **2026.2** | إضافة صلاحيات `Read App Content` و `Update App Content` للتحكم بالوصول للسكريبتات وسياسات SLA |

### نقاط الانتباه عند الترقية

- عند الترقية من إصدار أقدم من 2023.1: دور `Developer` **يبقى** في المشاريع الموجودة
- المشاريع الجديدة بعد الترقية تستخدم `Contributor` تلقائياً
- **لا يحدث تغيير تلقائي** للإسنادات الموجودة لمنع أي تأثير أمني

---

## 12. المصادر الرسمية

| المصدر | الرابط |
| -------- | -------- |
| صفحة إدارة الأدوار | [Roles](https://www.jetbrains.com/help/youtrack/cloud/manage-roles.html) |
| الأدوار الافتراضية | [Default Roles](https://www.jetbrains.com/help/youtrack/cloud/default-roles.html) |
| مقارنة صلاحيات الأدوار | [Permission Comparison for Default Roles](https://www.jetbrains.com/help/youtrack/cloud/permissions-comparison-for-default-roles.html) |
| إنشاء وتعديل الأدوار | [Create and Edit Roles](https://www.jetbrains.com/help/youtrack/cloud/create-and-edit-roles.html) |
| نسخ الأدوار | [Clone a Role](https://www.jetbrains.com/help/youtrack/cloud/clone-role.html) |
| دمج الأدوار | [Merge Roles](https://www.jetbrains.com/help/youtrack/cloud/merge-roles.html) |
| مرجع الصلاحيات | [Permissions Reference](https://www.jetbrains.com/help/youtrack/cloud/youtrack-permissions-reference.html) |
| إدارة الوصول | [Access Management](https://www.jetbrains.com/help/youtrack/cloud/Access-Management.html) |

---

> **ملاحظة:** هذا المستند مبني بالكامل على التوثيق الرسمي لـ YouTrack Cloud 2026.2 من JetBrains. جميع المعلومات مستخرجة من المصادر الرسمية المذكورة أعلاه.
