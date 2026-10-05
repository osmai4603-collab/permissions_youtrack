# المستويات التنظيمية المساعدة في نظام صلاحيات YouTrack

## بحث مستند إلى المصادر الرسمية لـ JetBrains

> **تاريخ البحث:** 2026-10-05
> **الإصدار المرجعي:** YouTrack Server / Cloud 2026.2
> **المصادر الرسمية المستخدمة:**
>
> - [Access Management](https://www.jetbrains.com/help/youtrack/server/access-management.html)
> - [Roles](https://www.jetbrains.com/help/youtrack/server/manage-roles.html)
> - [Default Roles](https://www.jetbrains.com/help/youtrack/server/default-roles.html)
> - [Permissions Reference](https://www.jetbrains.com/help/youtrack/server/youtrack-permissions-reference.html)
> - [Organizations](https://www.jetbrains.com/help/youtrack/server/organizations.html)
> - [Configure Access for Groups](https://www.jetbrains.com/help/youtrack/server/configure-access-for-a-user-group.html)

---

## 1. نظرة عامة على فلسفة الصلاحيات في YouTrack

يعتمد نظام الصلاحيات في YouTrack على مبدأ أساسي:

> **الصلاحيات لا تُمنح مباشرة للمستخدمين أبداً** — بل تُجمع في **أدوار (Roles)**، ثم تُسند الأدوار إلى المستخدمين أو المجموعات ضمن **نطاق (Scope)** محدد.

### المفاهيم الأساسية

| المفهوم | التعريف الرسمي |
| --------- | --------------- |
| **Permission** | تفويض يُمنح للمستخدم لأداء عملية معينة |
| **Role** | حاوية مسماة لمجموعة من الصلاحيات تُسند للمستخدمين أو المجموعات ضمن نطاق مشروع أو منظمة |
| **Scope** | النطاق الذي تُطبق فيه الصلاحية (عام، منظمة، مشروع) |
| **Group** | مجموعة من حسابات المستخدمين تُستخدم لإدارة العضوية وتعيين الأدوار بكفاءة |
| **Organization** | تسلسل هرمي يجمع المشاريع لتعيين أدوار الوصول بشكل موحد |
| **Project Team** | المستخدمون والمجموعات المرتبطون بمشروع؛ العضوية يمكن أن تمنح أدواراً |

---

## 2. نطاقات الصلاحيات (Permission Scopes) — المستويات التنظيمية الثلاثة

وفقاً للتوثيق الرسمي في صفحة [Permissions Reference](https://www.jetbrains.com/help/youtrack/server/youtrack-permissions-reference.html)، تُقسم الصلاحيات إلى ثلاثة نطاقات:

### 2.1 المستوى العام (Global Scope)

| الخاصية | الوصف |
| --------- | ------ |
| **التعريف** | صلاحيات تُمنح ضمن النطاق العام لـ YouTrack ولا تعتمد على مشروع محدد |
| **النطاق** | يشمل كامل نسخة YouTrack (Instance-wide) |
| **التعريف بالواجهة** | تُميز بشارة `Global` في قائمة الصلاحيات |
| **مثال** | لا يمكن منح صلاحية إنشاء مستخدمين في مشروع واحد فقط — بل في النطاق العام فقط |
| **مكان الإدارة** | Administration > Access Management |

**الصلاحيات المتوفرة فقط على المستوى العام:**

- `Create User` — إنشاء حسابات مستخدمين جديدة
- `Delete User` — حذف حسابات مستخدمين
- `Update User` — تعديل بيانات حسابات المستخدمين
- `Read User Basic` — عرض قائمة المستخدمين المسجلين
- `Read User Details` — عرض تفاصيل إضافية للمستخدمين
- `Update Self` — تعديل بيانات الحساب الشخصي
- `Low-level Admin Read` — قراءة الإعدادات الإدارية المنخفضة المستوى
- `Low-level Admin Write` — إدارة العمليات الإدارية المنخفضة المستوى
- `Create Project` — إنشاء مشاريع جديدة
- `Create Organization` — إنشاء منظمات جديدة

### 2.2 مستوى المنظمة (Organization Scope)

| الخاصية | الوصف |
| --------- | ------ |
| **التعريف** | صلاحيات محدودة بمنظمة معينة وتشمل جميع المشاريع المنتمية لها |
| **الغرض الرئيسي** | طريقة مريحة لمنح وصول موحد لجميع مشاريع منظمة ما |
| **الفائدة** | أكثر كفاءة من إدارة كل مشروع على حدة عند وجود مشاريع كثيرة تحت منظمة واحدة |
| **مكان الإدارة** | تبويب Access ضمن إعدادات المنظمة المحددة |

**الصلاحيات المتوفرة على مستوى المنظمة:**

- `Read Organization` — عرض المنظمات وسماتها
- `Update Organization` — تعديل سمات المنظمة وإدارة تعيينات المشاريع وحقوق الوصول
- `Delete Organization` — حذف سجلات المنظمة نهائياً

**كيف تعمل المنظمة في نظام الصلاحيات:**

وفقاً للتوثيق الرسمي:

> "Organizations let you add structure to your project management efforts by grouping resources and projects under one roof. They also give you a quick and easy way to manage uniform access rights to projects and project-related content."

**مثال عملي من التوثيق:**
> لديك 10 مشاريع وفريق تطوير واحد. تريد لهذا الفريق دوراً في جميع المشاريع. يمكنك إما تعيين دور للفريق في كل مشروع (10 مرات)، أو تجميع المشاريع كمنظمة ومنح الفريق دوراً في نطاق هذه المنظمة. مستقبلاً، إذا أنشأت مشروعاً جديداً يجب أن ينتمي إليه هذا الفريق، تضيف المشروع للمنظمة ويحصل الفريق تلقائياً على نفس الوصول.

### 2.3 مستوى المشروع (Project Scope)

| الخاصية | الوصف |
| --------- | ------ |
| **التعريف** | صلاحيات تسمح بعمليات متعلقة بمشروع محدد |
| **النطاق** | مشروع واحد أو أكثر بشكل فردي |
| **مثال** | دور يحتوي صلاحية `Read Project Basic` يمنح حق عرض خصائص ومحتوى مشروع معين فقط |
| **التوريث** | يمكن أن يرث المستخدمون الأدوار من خلال عضويتهم في فريق المشروع أو المجموعات |

**أهم الصلاحيات على مستوى المشروع:**

| الفئة | الصلاحيات |
| ------- | ---------- |
| **المشروع** | Read Project Basic, Read Project Full, Update Project, Delete Project |
| **المشكلات (Issues)** | Create Issue, Read Issue, Update Issue, Delete Issue, Link Issues, Read/Update Issue Private Fields |
| **التعليقات** | Create/Read/Update/Delete Issue Comment |
| **المرفقات** | Add/Update/Delete Attachment |
| **عناصر العمل** | Create/Read/Update Work Item |
| **المقالات** | Create/Read/Update/Delete Article |
| **التطبيقات** | Read/Update App Content |

---

## 3. آليات تعيين الصلاحيات (Assignment Mechanisms)

### 3.1 التعيين عبر المجموعات (Groups)

وفقاً للتوثيق الرسمي، الممارسة المعيارية هي تعيين الأدوار لـ **المجموعات** أو **فرق المشاريع** بدلاً من المستخدمين الأفراد:

```
المستخدم ← المجموعة ← الدور ← المشروع/المنظمة
```

- المستخدمون يرثون أدوار المجموعات التي ينتمون إليها
- يُبسط الإدارة بشكل كبير

### 3.2 التعيين عبر فريق المشروع (Project Team)

```
المستخدم ← فريق المشروع ← الدور الافتراضي للفريق
```

- الأدوار تُسند لأعضاء الفريق تلقائياً
- فريق المشروع يُحدد المستخدمين المتاحين كمكلفين (Assignees)

### 3.3 التعيين المباشر (Direct Assignment)

```
المستخدم ← الدور ← المشروع/المنظمة
```

- يمكن تعيين أدوار لحسابات مستخدمين فردية عند الحاجة لوصول غير معياري
- **غير مستحسن** كممارسة أساسية

---

## 4. الأدوار الافتراضية (Default Roles)

يوفر YouTrack ستة أدوار افتراضية مُعدة مسبقاً (للقراءة فقط — يمكن نسخها وتخصيصها):

### جدول الأدوار الافتراضية ونطاقاتها

| الدور | النطاق | الغرض |
| ------ | -------- | ------- |
| **System Admin** | عام (Global) | إدارة كاملة للنظام — يحتوي جميع الصلاحيات المتاحة |
| **Project Admin** | مشروع (Project) | إدارة المشاريع — نفس صلاحيات Contributor + إدارة إعدادات المشروع |
| **Contributor** | مشروع (Project) | العمل اليومي — إنشاء وتحديث المشكلات والتعليقات والمرفقات |
| **Observer** | عام (Global) | الوصول الأساسي — عرض المستخدمين وتحديث الحساب الشخصي |
| **User Manager** | عام (Global) | إدارة المستخدمين — إنشاء حسابات ودعوة مستخدمين |
| **Project Creator** | عام (Global) | إنشاء مشاريع — إنشاء مشاريع جديدة دون الوصول لمشاريع الآخرين |

### تفاصيل كل دور

#### System Admin

- يحتوي **جميع** الصلاحيات المتاحة في YouTrack
- مخصص للمسؤولين عن إدارة نسخة YouTrack

#### Project Admin

- جميع صلاحيات Contributor
- إضافة: `Read Project Full`, `Update Project`
- يستطيع إدارة Workflows, Apps, Integrations, Agile Boards على مستوى المشروع
- **لا يحتوي** صلاحيات عامة مثل `Create Project`, `Create User`

#### Contributor

- الصلاحيات الأساسية للعمل اليومي
- Issue: Read, Create, Update, Delete, Link, Read/Update Private Fields
- Comment: Create, Read, Update, Delete
- Attachment: Add, Update, Delete
- Work Item: Read, Update, Create
- Article: Read, Create
- **ملاحظة:** حل محل دور `Developer` منذ الإصدار 2023.1

#### Observer

- `Update Self` — تعديل الحساب الشخصي
- `Read User Basic` — عرض المستخدمين المسجلين
- `Read User Details` — عرض تفاصيل إضافية للمستخدمين
- يُمنح عادة على المستوى العام لجميع المستخدمين

#### User Manager

- `Create User` فقط على المستوى العام
- لمنح غير المسؤولين قدرة إضافة حسابات مستخدمين

#### Project Creator

- `Create Project` فقط على المستوى العام
- مالكو المشاريع يحصلون تلقائياً على دور Project Admin في مشاريعهم

---

## 5. آليات الصلاحيات الخاصة

### 5.1 الصلاحيات المتوارثة (Inherent Permissions)

بعض الصلاحيات تتوارث تلقائياً:

| إذا كان لديك | ترث تلقائياً |
| ------------- | ------------- |
| `Create Issue` | قراءة وتحديث الحقول العامة وإضافة الروابط **في المشكلات التي أنشأتها** |
| `Create Issue Comment` | قراءة تعليقاتك الخاصة |
| `Create Work Item` | قراءة عناصر عملك الخاصة |
| `Add Attachment` | تعديل وتقييد رؤية المرفقات التي أضفتها |
| `Create Article Comment` | قراءة وتحديث تعليقاتك على المقالات |

### 5.2 الصلاحيات الضمنية والتابعة (Implied & Dependent)

- عند إضافة صلاحية لها صلاحيات ضمنية، تُضاف الضمنية تلقائياً
- عند إزالة صلاحية لها صلاحيات تابعة، تُزال التابعة تلقائياً

**مثال:** إضافة `Read Issue` أو `Create Issue` → يُضاف `Read Project Basic` تلقائياً (لأنه من المستحيل عرض أو إنشاء مشكلات دون قراءة خصائص المشروع الأساسية).

---

## 6. مخطط الهيكل التنظيمي لنظام الصلاحيات

```
┌─────────────────────────────────────────────────────────────────┐
│                    YouTrack Instance                            │
│                   (المستوى العام - Global)                      │
│                                                                 │
│  صلاحيات: System, Users, Create Project, Create Organization   │
│  أدوار: System Admin, Observer, User Manager, Project Creator  │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              Organization (المنظمة)                       │  │
│  │                                                           │  │
│  │  صلاحيات: Read/Update/Delete Organization                │  │
│  │  الفائدة: وصول موحد لمجموعة مشاريع                      │  │
│  │                                                           │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │  │
│  │  │  Project A   │  │  Project B   │  │  Project C   │    │  │
│  │  │ (المشروع)    │  │ (المشروع)    │  │ (المشروع)    │    │  │
│  │  │              │  │              │  │              │    │  │
│  │  │ صلاحيات:     │  │ صلاحيات:     │  │ صلاحيات:     │    │  │
│  │  │ Issues       │  │ Issues       │  │ Issues       │    │  │
│  │  │ Comments     │  │ Comments     │  │ Comments     │    │  │
│  │  │ Attachments  │  │ Attachments  │  │ Attachments  │    │  │
│  │  │ Articles     │  │ Articles     │  │ Articles     │    │  │
│  │  │ Work Items   │  │ Work Items   │  │ Work Items   │    │  │
│  │  │ Apps         │  │ Apps         │  │ Apps         │    │  │
│  │  │ Project Mgmt │  │ Project Mgmt │  │ Project Mgmt │    │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘    │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────┐                               │
│  │  مشاريع خارج أي منظمة       │                               │
│  │  (تُدار فردياً)              │                               │
│  │  ┌──────────────┐            │                               │
│  │  │  Project D   │            │                               │
│  │  └──────────────┘            │                               │
│  └──────────────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 7. تدفق حل الصلاحيات (Resolution Flow)

```
المستخدم يطلب عملية ما
         │
         ▼
هل يملك دوراً عاماً يحتوي هذه الصلاحية؟
         │
    نعم ──► ✅ السماح
         │
    لا ──► هل ينتمي لمنظمة فيها دور يحتوي الصلاحية؟
                    │
               نعم ──► هل المشروع المستهدف ينتمي لهذه المنظمة؟
                              │
                         نعم ──► ✅ السماح
                              │
                         لا ──► ▼
                    │
               لا ──► هل يملك دوراً مباشراً في المشروع المستهدف؟
                              │
                         نعم ──► ✅ السماح
                              │
                         لا ──► هل يرث دوراً من مجموعة أو فريق المشروع؟
                                        │
                                   نعم ──► ✅ السماح
                                        │
                                   لا ──► ❌ الرفض
```

---

## 8. ملاحظات مهمة من التوثيق الرسمي

### 8.1 فصل تعريف الدور عن تعيينه

> YouTrack يفصل بين **تعريف الدور** (ما هي الصلاحيات التي يحتويها) و**تعيينه** (لمن وأين). عند تعيين دور، يُفعّل النظام فقط الصلاحيات المتوافقة مع نطاق التعيين.

**مثال:** إذا عُيّن دور يحتوي صلاحيات على مستوى المشروع في النطاق العام فقط، فلن تعمل صلاحيات المشروع.

### 8.2 Hub كنظام أساسي

في حالة ربط YouTrack Server بخدمة Hub خارجية، تُدار صلاحيات الوصول مباشرة في Hub. وهذا يشمل إدارة الوصول في الخدمات المتصلة مثل TeamCity.

### 8.3 الأدوار الافتراضية للقراءة فقط

الأدوار المُعدة مسبقاً في YouTrack هي **للقراءة فقط**. لتخصيصها، يجب **نسخها (Clone)** ثم تعديل النسخة.

### 8.4 الترتيب المُقترح لتهيئة الصلاحيات

وفقاً للتوثيق الرسمي، الترتيب المقترح:

1. إنشاء أدوار جديدة أو تهيئة الأدوار المحددة مسبقاً
2. إنشاء مجموعات جديدة أو تهيئة المجموعات المحددة مسبقاً
3. تعيين أدوار للمجموعات على أساس كل مشروع
4. إنشاء حسابات مستخدمين أو تمكين التسجيل الذاتي
5. تمكين وتهيئة وحدات المصادقة
6. تهيئة عضوية المجموعات للمستخدمين

---

## 9. ملخص المقارنة بين المستويات التنظيمية

| المعيار | المستوى العام (Global) | مستوى المنظمة (Organization) | مستوى المشروع (Project) |
| --------- | ---------------------- | --------------------------- | ---------------------- |
| **النطاق** | كامل النسخة | مشاريع المنظمة | مشروع واحد |
| **أنواع الصلاحيات** | System, Users, Create | Organization CRUD | Issues, Comments, Articles, etc. |
| **الأدوار الافتراضية** | System Admin, Observer, User Manager, Project Creator | (تُعيّن أدوار مخصصة) | Project Admin, Contributor |
| **حالة الاستخدام** | إدارة النظام الشامل | إدارة وصول موحد لمجموعة مشاريع | إدارة وصول مشروع محدد |
| **التوريث** | لا يرث — يُطبق مباشرة | يُوزع على جميع مشاريع المنظمة | محدود بالمشروع |
| **مكان الإدارة** | Administration > Access Management | Organization Settings > Access | Project Settings > Team/Access |

---

## 10. المراجع الرسمية الكاملة

| الصفحة | الرابط |
| -------- | -------- |
| Access Management | <https://www.jetbrains.com/help/youtrack/server/access-management.html> |
| Roles | <https://www.jetbrains.com/help/youtrack/server/manage-roles.html> |
| Default Roles | <https://www.jetbrains.com/help/youtrack/server/default-roles.html> |
| Permissions Reference | <https://www.jetbrains.com/help/youtrack/server/youtrack-permissions-reference.html> |
| Permission Comparison | <https://www.jetbrains.com/help/youtrack/server/permissions-comparison-for-default-roles.html> |
| Organizations | <https://www.jetbrains.com/help/youtrack/server/organizations.html> |
| Create Organization | <https://www.jetbrains.com/help/youtrack/server/create-organization.html> |
| Manage Organization Access | <https://www.jetbrains.com/help/youtrack/server/manage-organization-permissions.html> |
| Configure Group Access | <https://www.jetbrains.com/help/youtrack/server/configure-access-for-a-user-group.html> |
| Create and Edit Roles | <https://www.jetbrains.com/help/youtrack/server/create-and-edit-roles.html> |
