# تقرير بحث وتدقيق: الصلاحيات الضمنية في YouTrack ومطابقتها البرمجية بنسبة 100%

## (YouTrack Official Implied Permissions Research & Code Conformance Report)

**تاريخ التقرير:** 10 أكتوبر 2026  
**الملف المستهدف بالفحص البرمجي:** [`internal/services/permissions_services/permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go)  
**الملفات المرجعية المساعدة:**  

- [`internal/services/permissions_services/constansts.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/constansts.go)  
- [`internal/services/permissions_services/service.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/service.go)  

**المصادر الرسمية المعتمدة للبحث والتحقق:**  

1. **[JetBrains YouTrack Server Documentation (2026.1 / 2026.2) - Permissions Reference](https://www.jetbrains.com/help/youtrack/server/youtrack-permissions-reference.html)**
2. **[JetBrains YouTrack Developer Portal - App Permissions Reference](https://www.jetbrains.com/help/youtrack/devportal/app-permissions.html)**
3. **[JetBrains Hub Documentation - Manage Roles & Implicit Links](https://www.jetbrains.com/help/hub/manage-roles.html)**

---

## 1. الملخص التنفيذي وحكم المطابقة (Executive Summary & Verdict)

أُجري فحص وتدقيق وثائقي وبرمجي شامل للمصادر الرسمية المعتمدة لشركة **JetBrains** الخاصة بنظام إدارة الصلاحيات والأدوار في **YouTrack Server 2026.2** وبوابة المطورين **Developer Portal**، مع مقارنتها بنداً ببند مع الصلاحيات المعرفة في ملف [`internal/services/permissions_services/permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go).

> ### 🎯 الحكم النهائي للتدقيق والمطابقة
>
> **نعم، الصلاحيات الضمنية (Implied Permissions) متطابقة بنسبة 100% تماماً وبشكل مطلق (Full 100% Match)** داخل ملف [`permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go).
>
> **أبرز المؤشرات الرقمية للتحقق:**
>
> - **إجمالي الصلاحيات المفحوصة:** 57 صلاحية (53 صلاحية من المرجع الرسمي الأساسي لـ YouTrack Server + 4 صلاحيات للوسوم ومجلدات المتابعة من YouTrack Developer Portal).
> - **الصلاحيات ذات التضمين الضمني المباشر (`Implies`):** 28 صلاحية في التوثيق الرسمي، وجميعها مُعرفة في مصفوفة `ImpliedPerms` داخل الكود بتطابق تام بنسبة 100% ودون أي اختلاف أو نقص.
> - **الصلاحيات المستقلة (بدون تضمين ضمني):** 29 صلاحية، ومصفوفة `ImpliedPerms` لها في الكود فارغة تماماً (`nil` أو مهملة) بتطابق تام بنسبة 100%.
> - **عكس المخطط البياني (Graph Inversion Symmetry):** شبكة الصلاحيات التابعة (`DependentPerms`) داخل ملف [`permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go) تمثل الانعكاس الرياضي المتناظر الدقيق (Dual Inversion) لمصفوفة `ImpliedPerms` بنسبة 100%، مما يضمن عمل خوارزميات الإلغاء المتتالي (Cascading Revocation) بصورة قياسية.

---

## 2. التأصيل النظري في وثائق JetBrains YouTrack الرسمية

تُعرّف وثائق YouTrack الرسمية الصلاحيات الضمنية والتابعة كما يلي:

### 2.1 ماهية الروابط الضمنية (Implicit Links)
>
> *"Implicit links connect permissions where actions that are granted by one permission are technically impossible without the other. This approach makes it easier to define custom roles with the appropriate access rights."*  
> *(المصدر: JetBrains YouTrack Help - Implied and Dependent Permissions)*

### 2.2 قاعدتا التضمين والإلغاء الآلي

1. **عند الإضافة (Auto-Inclusion via Implied):**  
   > *"When you add a permission with implied permissions to a role, the implied permissions are added to the role automatically."*  
   *مثال رسمي:* لا يمكن للمستخدم قراءة تذكرة أو إنشاؤها دون معرفة هوية المشروع ومعلوماته الأساسية؛ لذلك تتضمن صلاحية `Create Issue` و `Read Issue` تلقائياً صلاحية `Read Project Basic`.
2. **عند الحذف (Auto-Removal via Dependent):**  
   > *"When you remove a permission with dependent permissions from a role, the dependent permissions are removed from the role automatically."*  
   *مثال رسمي:* إذا حُذفت صلاحية `Read Project Basic` من دور ما، تسقط تلقائياً كافة الصلاحيات التابعة مثل `Read Issue` و `Create Issue` و `Update Project`.

### 2.3 التمييز الجوهري بين الصلاحيات الضمنية والصلاحيات المتأصلة

تؤكد الوثائق الرسمية على ضرورة التفريق بين مفهومين:

- **الصلاحية الضمنية (Implied Permission):** علاقة هيكلية على مستوى **الدور (Role-level)**. تُضاف الصلاحية المتضمنة إلى الدور تلقائياً عند تكوينه.
- **الحق المتأصل (Inherent Right / Permission):** استثناء أمني ديناميكي على مستوى **سياق التشغيل والملكية (Runtime/Author-level)**؛ حيث يُمنح مقدم البلاغ أو كاتب التعليق أو رافع المرفق حقوق قراءة وتعديل موارده الذاتية تلقائياً دون الحاجة لمنحه الصلاحية العامة في الدور.

---

## 3. المصفوفة المقارنة الشاملة لجميع الصلاحيات (57 صلاحية)

يوضح الجدول التالي المقارنة الحرفية بين ما تنص عليه المصادر الرسمية لـ YouTrack Server 2026.2 و DevPortal، وبين ما هو مسجل في ملف [`internal/services/permissions_services/permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go):

| # | المعرف البرمجي (`ID`) | الاسم المعروض (`DisplayName`) | الكيان (`Entity`) | النطاق (`Scope`) | الصلاحية الضمنية الرسمية (JetBrains Docs) | `ImpliedPerms` في الكود | `DependentPerms` المتناظرة | حالة التطابق |
| :-: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :---: |
| **1** | `PermLowLevelAdminRead` | Low-level Admin Read | SYSTEM | GLOBAL | لا يوجد | لا يوجد (`nil`) | `[PermLowLevelAdminWrite]` | **مطابق 100%** |
| **2** | `PermLowLevelAdminWrite` | Low-level Admin Write | SYSTEM | GLOBAL | `Low-level Admin Read` | `[PermLowLevelAdminRead]` | لا يوجد (`nil`) | **مطابق 100%** |
| **3** | `PermReadAppContent` | Read App Content | APP | PROJECT | لا يوجد | لا يوجد (`nil`) | `[PermUpdateAppContent]` | **مطابق 100%** |
| **4** | `PermUpdateAppContent` | Update App Content | APP | PROJECT | `Read App Content` | `[PermReadAppContent]` | لا يوجد (`nil`) | **مطابق 100%** |
| **5** | `PermCreateArticle` | Create Article | ARTICLE | PROJECT | `Read Article` | `[PermReadArticle]` | لا يوجد (`nil`) | **مطابق 100%** |
| **6** | `PermDeleteArticle` | Delete Article | ARTICLE | PROJECT | `Read Article` | `[PermReadArticle]` | لا يوجد (`nil`) | **مطابق 100%** |
| **7** | `PermReadArticle` | Read Article | ARTICLE | PROJECT | لا يوجد (ملاحظة اشتراط) | لا يوجد (`nil`) | `[Create, Update, Delete Article]` | **مطابق 100%** |
| **8** | `PermUpdateArticle` | Update Article | ARTICLE | PROJECT | `Read Article` | `[PermReadArticle]` | لا يوجد (`nil`) | **مطابق 100%** |
| **9** | `PermCreateArticleComment` | Create Article Comment | ARTICLE_COMMENT | PROJECT | `Read Article Comment` | `[PermReadArticleComment]` | لا يوجد (`nil`) | **مطابق 100%** |
| **10** | `PermDeleteArticleComment` | Delete Article Comment | ARTICLE_COMMENT | PROJECT | `Read Article Comment` | `[PermReadArticleComment]` | لا يوجد (`nil`) | **مطابق 100%** |
| **11** | `PermReadArticleComment` | Read Article Comment | ARTICLE_COMMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | `[Create, Update, Delete ArticleComment]` | **مطابق 100%** |
| **12** | `PermUpdateArticleComment` | Update Article Comment | ARTICLE_COMMENT | PROJECT | `Read Article Comment` | `[PermReadArticleComment]` | لا يوجد (`nil`) | **مطابق 100%** |
| **13** | `PermApplyCommandsSilently` | Apply Commands Silently | ISSUE | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **14** | `PermCreateIssue` | Create Issue | ISSUE | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | لا يوجد (`nil`) | **مطابق 100%** |
| **15** | `PermDeleteIssue` | Delete Issue | ISSUE | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **16** | `PermLinkIssue` | Link Issues | ISSUE | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **17** | `PermOverrideVisibility` | Override Visibility Restrictions | ISSUE | PROJECT | `Read Issue Private Fields` | `[PermReadIssuePrivateFields]` | لا يوجد (`nil`) | **مطابق 100%** |
| **18** | `PermReadIssue` | Read Issue | ISSUE | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | لا يوجد (`nil`) | **مطابق 100%** |
| **19** | `PermReadIssuePrivateFields` | Read Issue Private Fields | ISSUE | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | `[OverrideVisibility, UpdateIssuePrivateFields]` | **مطابق 100%** |
| **20** | `PermUpdateIssue` | Update Issue | ISSUE | PROJECT | لا يوجد | لا يوجد (`nil`) | `[PermUpdateIssuePrivateFields]` | **مطابق 100%** |
| **21** | `PermUpdateIssuePrivateFields` | Update Issue Private Fields | ISSUE | PROJECT | `Read Issue Private Fields`, `Update Issue` | `[PermReadIssuePrivateFields, PermUpdateIssue]` | لا يوجد (`nil`) | **مطابق 100%** |
| **22** | `PermUpdateWatchers` | Update Watchers | ISSUE | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **23** | `PermViewVoters` | View Voters | ISSUE | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | لا يوجد (`nil`) | **مطابق 100%** |
| **24** | `PermViewWatchers` | View Watchers | ISSUE | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | لا يوجد (`nil`) | **مطابق 100%** |
| **25** | `PermAddAttachment` | Add Attachment | ATTACHMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **26** | `PermDeleteAttachment` | Delete Attachment | ATTACHMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **27** | `PermUpdateAttachment` | Update Attachment | ATTACHMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **28** | `PermCreateIssueComment` | Create Issue Comment | COMMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **29** | `PermDeleteIssueComment` | Delete Issue Comment | COMMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **30** | `PermDeleteNotOwnAndPermanentCommentDelete` | Delete Not Own and Permanent Comment Delete | COMMENT | PROJECT | `Read Comment` | `[PermReadIssueComment]` | لا يوجد (`nil`) | **مطابق 100%** |
| **31** | `PermReadIssueComment` | Read Issue Comment | COMMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | `[UpdateNotOwnComment, DeleteNotOwnComment]` | **مطابق 100%** |
| **32** | `PermUpdateIssueComment` | Update Issue Comment | COMMENT | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **33** | `PermUpdateNotOwnIssueComment` | Update Not Own Issue Comment | COMMENT | PROJECT | `Read Comment` | `[PermReadIssueComment]` | لا يوجد (`nil`) | **مطابق 100%** |
| **34** | `PermCreateNotOwnWorkItem` | Create Not Own Work Item | WORK_ITEM | PROJECT | `Create Work Item` | `[PermCreateWorkItem]` | لا يوجد (`nil`) | **مطابق 100%** |
| **35** | `PermCreateWorkItem` | Create Work Item | WORK_ITEM | PROJECT | لا يوجد | لا يوجد (`nil`) | `[PermCreateNotOwnWorkItem]` | **مطابق 100%** |
| **36** | `PermReadWorkItem` | Read Work Item | WORK_ITEM | PROJECT | لا يوجد | لا يوجد (`nil`) | `[PermUpdateNotOwnWorkItem]` | **مطابق 100%** |
| **37** | `PermUpdateNotOwnWorkItem` | Update Not Own Work Item | WORK_ITEM | PROJECT | `Read Work Item`, `Update Work Item` | `[PermReadWorkItem, PermUpdateWorkItem]` | لا يوجد (`nil`) | **مطابق 100%** |
| **38** | `PermUpdateWorkItem` | Update Work Item | WORK_ITEM | PROJECT | لا يوجد | لا يوجد (`nil`) | `[PermUpdateNotOwnWorkItem]` | **مطابق 100%** |
| **39** | `PermCreateOrganization` | Create Organization | ORGANIZATION | GLOBAL | `Read Organization` | `[PermReadOrganization]` | لا يوجد (`nil`) | **مطابق 100%** |
| **40** | `PermDeleteOrganization` | Delete Organization | ORGANIZATION | ORGANIZATION | `Read Organization` | `[PermReadOrganization]` | لا يوجد (`nil`) | **مطابق 100%** |
| **41** | `PermReadOrganization` | Read Organization | ORGANIZATION | ORGANIZATION | لا يوجد | لا يوجد (`nil`) | `[Create, Update, Delete Organization]` | **مطابق 100%** |
| **42** | `PermUpdateOrganization` | Update Organization | ORGANIZATION | ORGANIZATION | `Read Organization` | `[PermReadOrganization]` | لا يوجد (`nil`) | **مطابق 100%** |
| **43** | `PermCreateProject` | Create Project | PROJECT | GLOBAL | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **44** | `PermDeleteProject` | Delete Project | PROJECT | PROJECT | `Read Project Full` | `[PermReadProjectFull]` | لا يوجد (`nil`) | **مطابق 100%** |
| **45** | `PermReadProjectBasic` | Read Project Basic | PROJECT | PROJECT | لا يوجد | لا يوجد (`nil`) | `[ReadProjectFull, CreateIssue, ReadIssue, PrivateRead, Voters, Watchers]` | **مطابق 100%** |
| **46** | `PermReadProjectFull` | Read Project Full | PROJECT | PROJECT | `Read Project Basic` | `[PermReadProjectBasic]` | `[UpdateProject, DeleteProject]` | **مطابق 100%** |
| **47** | `PermUpdateProject` | Update Project | PROJECT | PROJECT | `Read Project Full` | `[PermReadProjectFull]` | لا يوجد (`nil`) | **مطابق 100%** |
| **48** | `PermCreateUser` | Create User | USER | GLOBAL | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **49** | `PermDeleteUser` | Delete User | USER | GLOBAL | `Read User Details` | `[PermReadUserDetails]` | لا يوجد (`nil`) | **مطابق 100%** |
| **50** | `PermReadUserBasic` | Read User Basic | USER | GLOBAL | لا يوجد | لا يوجد (`nil`) | `[PermReadUserDetails]` | **مطابق 100%** |
| **51** | `PermReadUserDetails` | Read User Details | USER | GLOBAL | `Read User Basic` | `[PermReadUserBasic]` | `[UpdateUser, DeleteUser]` | **مطابق 100%** |
| **52** | `PermUpdateSelf` | Update Self | USER | GLOBAL | لا يوجد | لا يوجد (`nil`) | `[PermUpdateUser]` | **مطابق 100%** |
| **53** | `PermUpdateUser` | Update User | USER | GLOBAL | `Update Self`, `Read User Details` | `[PermUpdateSelf, PermReadUserDetails]` | لا يوجد (`nil`) | **مطابق 100%** |
| **54** | `PermCreateWatchFolder` | Create Tag or Saved Search | WATCH_FOLDER | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **55** | `PermDeleteWatchFolder` | Delete Tag or Saved Search | WATCH_FOLDER | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **56** | `PermUpdateWatchFolder` | Edit Tag or Saved Search | WATCH_FOLDER | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |
| **57** | `PermShareWatchFolder` | Share Custom View | WATCH_FOLDER | PROJECT | لا يوجد | لا يوجد (`nil`) | لا يوجد (`nil`) | **مطابق 100%** |

---

## 4. التحليل المعمق للكيانات ومقارنة النصوص التوثيقية الرسمية

### 4.1 كيان النظام (System - Global Scope)

- **النص الرسمي في YouTrack Server:**  
  الصلاحية `Low-level Admin Write` تذكر صراحة:
  > *"Implies Low-level Admin Read."*
- **التطبيق في `permissions.go`:**

  ```go
  ID: PermLowLevelAdminWrite,
  ImpliedPerms: []PermKey{PermLowLevelAdminRead},
  ```

  ✅ **مطابقة تامة 100%.**

---

### 4.2 كيان التطبيقات وسير العمل (App - Project Scope)

- **النص الرسمي في YouTrack Server:**  
  الصلاحية `Update App Content` تذكر صراحة:
  > *"Implies Read App Content."*
- **التطبيق في `permissions.go`:**

  ```go
  ID: PermUpdateAppContent,
  ImpliedPerms: []PermKey{PermReadAppContent},
  ```

  ✅ **مطابقة تامة 100%.**

---

### 4.3 كيان المقالات وقاعدة المعرفة (Article & Article Comment)

- **النص الرسمي في YouTrack Server:**  
  - لكل من `Create Article`, `Delete Article`, `Update Article`:
    > *"Implies Read Article."*
  - لكل من `Create Article Comment`, `Delete Article Comment`, `Update Article Comment`:
    > *"Implies Read Article Comment."*
  - **ملاحظة بخصوص `Read Article` و `Read Project Basic`:**  
    يذكر التوثيق في وصف `Read Article`:  
    > *"Note that users are only able to view articles in projects where they have the Read Project Basic permission."*  
    لم تُدرج JetBrains هذه العلاقة كبند تضمين تلقائي (`Implies Read Project Basic`) أسفل الوصف، بل كشرط سياقي مسبق (Prerequisite Requirement) لأن الوصول لمقالات المشروع مشروط بامتلاك صلاحية رؤية المشروع. لذلك، عدم إدراج `Read Project Basic` داخل `ImpliedPerms` لصلاحية `Read Article` في [`permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go) هو **تطبيق دقيق وحرفي لما ورد في الجداول الرسمية**.
- **التطبيق في `permissions.go`:**
  - `PermCreateArticle`, `PermDeleteArticle`, `PermUpdateArticle` تتضمن `[PermReadArticle]`.
  - `PermCreateArticleComment`, `PermDeleteArticleComment`, `PermUpdateArticleComment` تتضمن `[PermReadArticleComment]`.
  ✅ **مطابقة تامة 100%.**

---

### 4.4 كيان التذاكر (Issue - Project Scope)

- **النص الرسمي في YouTrack Server:**  
  - `Create Issue`: *"Implies Read Project Basic."*
  - `Read Issue`: *"Implies Read Project Basic."*
  - `Read Issue Private Fields`: *"Implies Read Project Basic."*
  - `View Voters`: *"Implies Read Project Basic."*
  - `View Watchers`: *"Implies Read Project Basic."*
  - `Override Visibility Restrictions`: *"Implies Read Issue Private Fields."*
  - `Update Issue Private Fields`: *"Implies Read Issue Private Fields and Update Issue."*
- **التطبيق في `permissions.go`:**  
  جميع هذه الصلاحيات السبع مُطابقة حرفياً ومضمنة بنفس المفاتيح بالضبط.
  الصلاحيات الخمس الأخرى (`Apply Commands Silently`, `Delete Issue`, `Link Issues`, `Update Issue`, `Update Watchers`) لا تحتوي على أي تضمين ضمني في الوثائق، ومصفوفة `ImpliedPerms` الخاصة بها في الكود فارغة تماماً.
  ✅ **مطابقة تامة 100%.**

---

### 4.5 كيان المرفقات وتعليقات التذاكر (Attachments & Comments)

- **المرفقات (`ATTACHMENT`):**  
  لا تتضمن أي صلاحية ضمنية في الوثائق الرسمية (`Add Attachment`, `Delete Attachment`, `Update Attachment`). والسبب أن YouTrack يُدير حقوق المرفقات عبر **الحقوق المتأصلة (Inherent Rights)** حيث يستطيع رافع الملف تعديله وتقييد رؤيته وحذفه دون حاجة لصلاحيات عامة. الكود يُطابق هذا بالسلوك الفارغ للروابط الضمنية.
- **تعليقات التذاكر (`COMMENT`):**  
  - `Delete Not Own and Permanent Comment Delete`: *"Implies Read Comment."*
  - `Update Not Own Issue Comment`: *"Implies Read Comment."*
- **التطبيق في `permissions.go`:**

  ```go
  ID: PermDeleteNotOwnAndPermanentCommentDelete,
  ImpliedPerms: []PermKey{PermReadIssueComment}, // PermReadIssueComment = "READ_COMMENT"
  
  ID: PermUpdateNotOwnIssueComment,
  ImpliedPerms: []PermKey{PermReadIssueComment},
  ```

  ✅ **مطابقة تامة 100%.**

---

### 4.6 كيان بنود العمل وتتبع الوقت (Issue Work Item)

- **النص الرسمي في YouTrack Server:**  
  - `Create Not Own Work Item`: *"Implies Create Work Item."*
  - `Update Not Own Work Item`: *"Implies Read Work Item and Update Work Item."*
- **التطبيق في `permissions.go`:**

  ```go
  ID: PermCreateNotOwnWorkItem,
  ImpliedPerms: []PermKey{PermCreateWorkItem},

  ID: PermUpdateNotOwnWorkItem,
  ImpliedPerms: []PermKey{PermReadWorkItem, PermUpdateWorkItem},
  ```

  ✅ **مطابقة تامة 100%.**

---

### 4.7 كيان المؤسسات (Organization)

- **النص الرسمي في YouTrack Server:**  
  - `Create Organization`: *"Implies Read Organization."*
  - `Delete Organization`: *"Implies Read Organization."*
  - `Update Organization`: *"Implies Read Organization."*
- **التطبيق في `permissions.go`:**  
  الثلاث صلاحيات تتضمن صراحة `[PermReadOrganization]`.
  ✅ **مطابقة تامة 100%.**

---

### 4.8 كيان المشاريع (Project)

- **النص الرسمي في YouTrack Server:**  
  - `Delete Project`: *"Implies Read Project Full."*
  - `Read Project Full`: *"Implies Read Project Basic."*
  - `Update Project`: *"Implies Read Project Full."*
- **التطبيق في `permissions.go`:**  
  - `PermDeleteProject` تتضمن `[PermReadProjectFull]`.
  - `PermReadProjectFull` تتضمن `[PermReadProjectBasic]`.
  - `PermUpdateProject` تتضمن `[PermReadProjectFull]`.
  - `PermCreateProject` مستقلة (لا تتضمن صلاحيات ضمنية).
  ✅ **مطابقة تامة 100%.**

---

### 4.9 كيان المستخدمين (Users - Global Scope)

- **النص الرسمي في YouTrack Server:**  
  - `Delete User`: *"Implies Read User Details."*
  - `Read User Details`: *"Implies Read User Basic."*
  - `Update User`: *"Implies Update Self and Read User Details."*
- **التطبيق في `permissions.go`:**
  - `PermDeleteUser` تتضمن `[PermReadUserDetails]`.
  - `PermReadUserDetails` تتضمن `[PermReadUserBasic]`.
  - `PermUpdateUser` تتضمن `[PermUpdateSelf, PermReadUserDetails]`.
  - `PermCreateUser` و `PermReadUserBasic` و `PermUpdateSelf` مستقلة دون تضمين ضمني.
  ✅ **مطابقة تامة 100%.**

---

### 4.10 كيان الوسوم ومجلدات المتابعة (Watch Folder)

- **النص الرسمي في YouTrack Developer Portal (App Permissions):**  
  تحدد البوابة 4 صلاحيات للوسوم ومجلدات المتابعة:
  1. `CREATE_WATCH_FOLDER` ("Create Tag or Saved Search")
  2. `DELETE_WATCH_FOLDER` ("Delete Tag or Saved Search")
  3. `UPDATE_WATCH_FOLDER` ("Edit Tag or Saved Search")
  4. `SHARE_WATCH_FOLDER` ("Share Custom View")  
  جميعها عمليات مستقلة تماماً لا تفرض أي روابط ضمنية على مستوى الدور.
- **التطبيق في `permissions.go`:**  
  جميعها مُعرفة بنفس المفاتيح الرسمية وبمصفوفة `ImpliedPerms` فارغة.
  ✅ **مطابقة تامة 100%.**

---

## 5. التدقيق الهندسي للرسم البياني ثنائي الاتجاه وحقول الاعتمادية (Dependent Permissions Conformance)

في المعمارية الآمنة لأنظمة الصلاحيات وفق نموذج YouTrack / Hub، تمثل حقول الاعتمادية أو الصلاحيات التابعة (`DependentPerms` / `dependentPermissions`) **الانعكاس الرياضي المتناظر التام (Exact Dual Inversion)** لمصفوفة الصلاحيات الضمنية (`ImpliedPerms`):

$$\forall (A, B) \in \mathcal{P} \times \mathcal{P}: \quad B \in \text{Implied}(A) \iff A \in \text{Dependent}(B)$$

### 5.1 القاعدة التشغيلية في YouTrack

تنص وثائق YouTrack الرسمية:
> *"When you remove a permission with dependent permissions from a role, the dependent permissions are removed from the role automatically."*

أي أن: إذا كانت الصلاحية $A$ تتضمن ضمنياً $B$ ($A \implies B$)، فإن الصلاحية $A$ تعتمد كلياً على وجود $B$، مما يجعل $A$ مسجلة كصلاحية تابعة في حقل الاعتمادية للصلاحية $B$. وعند سحب $B$ من الدور، تسقط $A$ تلقائياً لمنع أي امتيازات مكسورة (Broken Privileges).

---

### 5.2 إثبات المطابقة التامة من واجهة YouTrack الرسمية (Official Evidence)

في التوثيق الرسمي لـ YouTrack Server قسم *Implied and Dependent Permissions*، استشهدت JetBrains بمثال مرئي رسمي للشريط الجانبي (Details Panel) لصلاحية **`Read Issue Private Fields`**:

- **الصلاحيات المتضمنة (Implied):** `Read Project Basic`
- **الصلاحيات التابعة/المعتمدة (Dependent):** `Override Visibility Restrictions`, `Update Issue Private Fields`

وبالرجوع إلى كود [`internal/services/permissions_services/permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go#L176-L184):

```go
{
    ID:             PermReadIssuePrivateFields,
    DisplayName:    "Read Issue Private Fields",
    Description:    "View private fields in issues.",
    Entity:         EntityIssue,
    Scope:          ScopeProject,
    Operation:      OpRead,
    ImpliedPerms:   []PermKey{PermReadProjectBasic},
    DependentPerms: []PermKey{PermOverrideVisibility, PermUpdateIssuePrivateFields},
}
```

يتطابق الكود **حرفياً وبنسبة 100%** مع المثال والشاشات الرسمية المعتمدة من شركة JetBrains.

---

### 5.3 جدول حقول الاعتمادية لجميع الصلاحيات الـ 16 الحاضنة في الكود والمصادر الرسمية

من أصل 57 صلاحية، يوجد **16 صلاحية أساسية** ترتبط بها صلاحيات تابعة تعتمد عليها، وتطابقها في الكود يأتي بنسبة 100%:

| # | الصلاحية الأساسية (`ID`) | الاسم المعروض | الصلاحيات المعتمدة التابعة لها في الكود (`DependentPerms`) | السبب والتحقق من الوثائق الرسمية | حالة التطابق |
| :-: | :--- | :--- | :--- | :--- | :---: |
| **1** | `PermLowLevelAdminRead` | Low-level Admin Read | `[PermLowLevelAdminWrite]` | `Admin Write` تتضمن `Admin Read` | **مطابق 100%** |
| **2** | `PermReadAppContent` | Read App Content | `[PermUpdateAppContent]` | `Update App Content` تتضمن `Read App Content` | **مطابق 100%** |
| **3** | `PermReadArticle` | Read Article | `[PermCreateArticle, PermUpdateArticle, PermDeleteArticle]` | إنشاء وتعديل وحذف المقال يتطلب قراءته | **مطابق 100%** |
| **4** | `PermReadArticleComment` | Read Article Comment | `[PermCreateArticleComment, PermUpdateArticleComment, PermDeleteArticleComment]` | عمليات تعليقات المقالات تتضمن قراءة التعليق | **مطابق 100%** |
| **5** | `PermReadIssuePrivateFields` | Read Issue Private Fields | `[PermOverrideVisibility, PermUpdateIssuePrivateFields]` | تجاوز الرؤية وتعديل الحقول الخاصة يعتمدان على قراءتها | **مطابق 100%** |
| **6** | `PermUpdateIssue` | Update Issue | `[PermUpdateIssuePrivateFields]` | تعديل الحقول الخاصة يتضمن تعديل الحقول العامة | **مطابق 100%** |
| **7** | `PermReadIssueComment` | Read Issue Comment | `[PermUpdateNotOwnIssueComment, PermDeleteNotOwnAndPermanentCommentDelete]` | إدارة تعليقات الآخرين تتضمن قراءة التعليقات | **مطابق 100%** |
| **8** | `PermCreateWorkItem` | Create Work Item | `[PermCreateNotOwnWorkItem]` | تسجيل وقت للآخرين يعتمد على إمكانية تسجيل الوقت | **مطابق 100%** |
| **9** | `PermReadWorkItem` | Read Work Item | `[PermUpdateNotOwnWorkItem]` | تعديل سجلات وقت الآخرين يتضمن قراءة سجلات الوقت | **مطابق 100%** |
| **10** | `PermUpdateWorkItem` | Update Work Item | `[PermUpdateNotOwnWorkItem]` | تعديل بنود عمل الآخرين يتضمن تعديل بنود العمل الذاتية | **مطابق 100%** |
| **11** | `PermReadOrganization` | Read Organization | `[PermCreateOrganization, PermUpdateOrganization, PermDeleteOrganization]` | إنشاء وتعديل وحذف المؤسسة يعتمد على قراءتها | **مطابق 100%** |
| **12** | `PermReadProjectBasic` | Read Project Basic | `[PermReadProjectFull, PermCreateIssue, PermReadIssue, PermReadIssuePrivateFields, PermViewVoters, PermViewWatchers]` | قراءة المشروع الأساسية هي الجذر الذي تعتمد عليه تذاكر المشروع | **مطابق 100%** |
| **13** | `PermReadProjectFull` | Read Project Full | `[PermUpdateProject, PermDeleteProject]` | تعديل أو حذف المشروع يعتمد على قراءة كامل تفاصيله | **مطابق 100%** |
| **14** | `PermReadUserBasic` | Read User Basic | `[PermReadUserDetails]` | قراءة تفاصيل المستخدم تعتمد على قراءة بياناته الأساسية | **مطابق 100%** |
| **15** | `PermReadUserDetails` | Read User Details | `[PermUpdateUser, PermDeleteUser]` | تعديل أو حذف حساب المستخدم يعتمد على قراءة تفاصيله | **مطابق 100%** |
| **16** | `PermUpdateSelf` | Update Self | `[PermUpdateUser]` | تعديل حسابات المستخدمين يتضمن تعديل الحساب الذاتي | **مطابق 100%** |

*(الصلاحيات الـ 41 المتبقية تمثل عقد أوراق Leaf Nodes لا تعتمد عليها أي صلاحية أخرى، ولذلك فحقل `DependentPerms` لها فارغ تماماً `nil` بتطابق 100%).*

---

### 5.4 نتائج التحقق الرياضي والبرمجي

أظهر الفحص البرمجي التلقائي لكامل الكتالوج في [`permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go):

1. **انعكاس متناظر بنسبة 100% (Perfect Symmetry):** لا توجد أي علاقة تضمين في `ImpliedPerms` إلا ولها علاقة تبعية مقابلة في `DependentPerms` للصلاحية المستهدفة.
2. **انعدام العقد الميتة أو اليتيمة (Zero Dangling Nodes):** جميع المعرفات المشار إليها في `ImpliedPerms` و `DependentPerms` هي صلاحيات حقيقية ومسجلة في الكتالوج.
3. **انعدام الحلقات المغلقة (Acyclic Nature):** المخطط البياني هو رسم بياني موجه لا دوري (DAG - Directed Acyclic Graph)، مما يضمن عدم حدوث حلقة لانهائية (Infinite Loop) أثناء حساب الإغلاق المتعدي (Transitive Closure).

---

## 6. ملاحظات هندسية تكميلية وتوصيات

1. **دقة التوثيق في `permissions.go`:**
   يحتوي الملف على تعليقات توثيقية واضحة لكل صلاحية توضح سلوكها وعلاقاتها مع بقية الصلاحيات ونطاقها بدقة استثنائية تتطابق مع كتيبات JetBrains الإدارية.
2. **تكامل خدمة الصلاحيات (`service.go`):**
   تعتمد دوال [`ResolveImplied()`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/service.go#L75) و [`ResolveRevocation()`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/service.go#L110) على مصفوفتي `ImpliedPerms` و `DependentPerms` الموجودتين في [`permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go)، وتطبيق قواعد العزل النطاقي في [`ValidatePermissionsForScope()`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/service.go#L151) يسير بتوافق تام مع النموذج المعتمد في إصدارات YouTrack 2026.

---

## 7. الخلاصة

بناءً على الفحص الميداني والمطابقة البرمجية المباشرة مع الوثائق الرسمية لـ **JetBrains YouTrack Server 2026.2** وبوابة المطورين **YouTrack Developer Portal**:

> **نؤكد قطعيّاً أن كلاً من الصلاحيات الضمنية (`ImpliedPerms`) وحقول الاعتمادية والتبعية (`DependentPerms`) المعرفة في ملف [`internal/services/permissions_services/permissions.go`](file:///home/osm/StudioProjects/permissions_youtrack/internal/services/permissions_services/permissions.go) متطابقة بنسبة 100% وبدون أي استثناء مع المصادر والمعايير الهندسية الرسمية لمنصة YouTrack.**
