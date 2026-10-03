# الخطة التنفيذية لتجريد خدمة الأدوار (Roles Service Decoupling Plan)

> **تاريخ الإعداد:** 3 أكتوبر 2026  
> **حالة الوثيقة:** معتمدة للتنفيذ  
> **الهدف الأساسي:** تحويل `roles_service` إلى محرك تحكم بالوصول مبني على الأدوار (Pure Generic RBAC Engine) مجرد بالكامل من أي كينونات خارجية ملموسة مثل (`Group`, `Project`, `Organization`, `Issue`, `Article`).

---

## جدول المحتويات

1. [المقدمة والرؤية المعمارية](#1-المقدمة-والرؤية-المعمارية)
2. [تحليل الوضع الحالي ومواضع الارتباط الوثيق (Coupling Analysis)](#2-تحليل-الوضع-الحالي-ومواضع-الارتباط-الوثيق-coupling-analysis)
3. [التصميم المعماري المجرد المستهدف (Target Architecture)](#3-التصميم-المعماري-المجرد-المستهدف-target-architecture)
4. [الواجهات ونماذج البيانات التجريدية (SPI & Abstract Models)](#4-الواجهات-ونماذج-البيانات-التجريدية-spi--abstract-models)
5. [مراحل الخطة التنفيذية خطوة بخطوة (Execution Phases)](#5-مراحل-الخطة-التنفيذية-خطوة-بخطوة-execution-phases)
6. [مصفوفة المقارنة المعمارية (Before vs. After)](#6-مصفوفة-المقارنة-المعمارية-before-vs-after)
7. [خطة الاختبار والتحقق من الجودة (Testing & QA Plan)](#7-خطة-الاختبار-والتحقق-من-الجودة-testing--qa-plan)

---

## 1. المقدمة والرؤية المعمارية

في الأنظمة البرمجية الحديثة وفقاً لمبادئ **Clean Architecture** و **Domain-Driven Design (DDD)**، يجب أن تكون خدمة الأدوار (`roles_service`) محركاً مستقلاً (Generic RBAC Engine) يركز على مبدأ المسؤولية الواحدة (**Single Responsibility Principle**):

* **ما تملكه خدمة الأدوار:**
  * إدارة تعريفات الأدوار وحاويات الصلاحيات (`Roles`).
  * دورة حياة الدور (إنشاء، تعديل، حذف، نسخ، دمج).
  * تعيين دور لهوية ما ضمن نطاق معين (`Role Assignments`).
  * تجميع وحساب الصلاحيات الفعالة (`Effective Permissions Calculation`).

* **ما يجب أن يُعزل خارج خدمة الأدوار:**
  * كينونات المجموعات والهياكل الإدارية (`Group`, `UserGroups`).
  * كينونات المشاريع والمنظمات (`Project`, `Organization`).
  * منطق فِرق المشاريع (`Project Teams`).
  * الكيانات التشغيلية والسياقية الخاصة بالتطبيقات (`Issue`, `Article`, `CustomFields`).
  * قواعد الوصول المعتمدة على محتوى الكينونة أو علاقة المستخدم بها (مثل: محرر المقال، مقدم التذكرة).

---

## 2. تحليل الوضع الحالي ومواضع الارتباط الوثيق (Coupling Analysis)

تتضمن ملفات `internal/services/roles_service` حالياً نقاط ارتباط مباشرة تعيق التجريد وقابلية إعادة الاستخدام:

```
roles_service (الحالي)
 ├── assignment.go   ──> يحتوي على struct Group, struct Project, struct Organization
 ├── service.go      ──> يخزن خرائط مباشرة للكيانات: svc.projects, svc.organizations, svc.groups
 │                   ──> دوال إدارية: RegisterProject, RegisterGroup, AddProjectTeamMember...
 ├── aggregation.go  ──> يقرأ مباشرة svc.groups و svc.projects لحساب الانتماء والشجرة
 └── visibility.go   ──> يحتوي على struct IssueContext, struct ArticleContext
                     ──> دوال فحص محددة: CanViewIssue, CanViewArticle, CanAccessIssuePrivateField
```

### تفصيل مشكلات الوضع الحالي

1. **تلوث واجهة الخدمة (`IRolesService`)**: تحتوي الواجهة على 10 دوال لإدارة مشاريع ومجموعات ومنظمات، وهي اختصاصات لخدمات أخرى.
2. **تخزين حالات لا تخص الأدوار**: الخدمة تخزن بيانات المشاريع والمنظمات داخل ذاكرتها وحمايتها بأقفال (`sync.RWMutex`) مشتركة.
3. **صعوبة التبديل والتوسع**: لا يمكن استخدام خدمة الأدوار مع نطاقات أخرى (مثل المستودعات، لوحات المهام، مساحات العمل) دون تعديل كود الخدمة الداخلي.
4. **خلط طبقة السياسات بنواة الصلاحيات**: تقييم التذاكر والمقالات ينتمي لطبقة ترخيص التطبيق (`Application Policy Layer`) وليس لطبقة إدارة الأدوار.

---

## 3. التصميم المعماري المجرد المستهدف (Target Architecture)

```
┌────────────────────────────────────────────────────────────────────────┐
│                   Contextual Policy & Access Layer                     │
│                 (internal/services/policy_service)                     │
│  - IssueAccessPolicy (CanViewIssue, FieldAccess)                       │
│  - ArticleAccessPolicy (Hierarchical Articles)                         │
│  - InherentRightsEvaluator                                             │
└──────────────────────────────────┬─────────────────────────────────────┘
                                   │ Uses
                                   ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Generic Roles Service (Pure RBAC)                    │
│                 (internal/services/roles_service)                      │
│                                                                        │
│   - Role Lifecycle (CRUD, Clone, Merge)                                │
│   - Generic Assignment (PrincipalID ↔ RoleID @ ScopeRef)              │
│   - Effective Permissions Engine (via Providers)                       │
└──────────────┬──────────────────────────────────────────┬──────────────┘
               │ Uses SPI                                 │ Uses SPI
               ▼                                          ▼
┌──────────────────────────────┐        ┌──────────────────────────────┐
│  PrincipalHierarchyProvider  │        │    ScopeHierarchyProvider    │
│  (Resolves groups/identities)│        │   (Resolves parent scopes)   │
└──────────────────────────────┘        └──────────────────────────────┘
```

---

## 4. الواجهات ونماذج البيانات التجريدية (SPI & Abstract Models)

### أ. النطاق المجرد (`ScopeRef`)

بدلاً من حقول نصية منفصلة مثل `ProjectID` أو `OrganizationID`:

```go
// ScopeRef يحدد أي نطاق في النظام بشكل عام ومجرد
type ScopeRef struct {
    Level    perms.ScopeLevel `json:"level"`     // GLOBAL, ORGANIZATION, PROJECT (أو أي مستوى مستقبلي)
    TargetID string           `json:"target_id"` // المعرّف الفريد للهدف (فارغ في حالة GLOBAL)
}

func GlobalScope() ScopeRef {
    return ScopeRef{Level: perms.ScopeGlobal, TargetID: ""}
}

func ScopedTarget(level perms.ScopeLevel, targetID string) ScopeRef {
    return ScopeRef{Level: level, TargetID: targetID}
}
```

### ب. التعيين المجرد للأدوار (`RoleAssignment`) وتحليل الحقول

في الكود السابق، كان نموذج التعيين كالتالي:

```go
// النموذج السابق (قبل التجريد)
type RoleAssignment struct {
    ID            string           `json:"id"`
    RoleID        string           `json:"role_id"`
    PrincipalType PrincipalType    `json:"principal_type"` // USER or GROUP
    PrincipalID   string           `json:"principal_id"`
    Scope         perms.ScopeLevel `json:"scope"`          // حقلان منفصلان قد يؤديان إلى حالات غير متناسقة
    TargetID      string           `json:"target_id"`
}
```

أما في النموذج المستهدف المعتمد (مع الاحتفاظ بـ `PrincipalType` ودمج النطاق في `ScopeRef`):

```go
// أنواع الفاعلين (قابلة للتوسع دون ربط الخدمة بكينونات داخلية)
type PrincipalType string

const (
    PrincipalUser  PrincipalType = "USER"
    PrincipalGroup PrincipalType = "GROUP"
)

// النموذج المعتمد
type RoleAssignment struct {
    ID            string        `json:"id"`
    RoleID        string        `json:"role_id"`
    PrincipalType PrincipalType `json:"principal_type"` // تصنيف نوع الفاعل (USER أو GROUP أو أنواع مستقبلية)
    PrincipalID   string        `json:"principal_id"`   // مُعرّف الفاعل
    Scope         ScopeRef      `json:"scope"`          // كائن قيمة متكامل ومحصن ضد التناقض
}
```

#### أسباب وتفاصيل هندسة الحقول

1. **الاحتفاظ بحقل `PrincipalType` (`USER` / `GROUP`):**
   * **التصنيف والفلترة السريعة (Explicit Classification & Querying):** يتيح للخدمة وواجهات الـ API تصفية التعيينات حسب نوع الفاعل (مثلاً: استعراض جميع تعيينات المجموعات فقط لمشروع معين، أو تعيينات المستخدمين المباشرة) دون الحاجة لاستدعاء خدمات الهوية الخارجية للتحقق من هوية كل معرّف.
   * **سجل التدقيق والتتبع (Auditability & Metadata):** يوفر وضوحاً كاملاً في سجلات النظام وتقارير الصلاحيات حول كيفية منح الدور (هل مُنح للمستخدم بصفته فرداً، أم لمجموعة).
   * **الحفاظ على التجريد الكامل:** وجود `PrincipalType` كـ "وسم تصنيفي" (Discriminator/Tag) **لا يكسر التجريد**، لأن خدمة الأدوار لن تحتوي على جداول المستخدمين أو المجموعات ولا تدير أعضاءها، بل تعتمد في حل العلاقات الهرمية على مزود الهويات الخارجي (`PrincipalHierarchyProvider`).

2. **دمج حقلي `Scope` و `TargetID` في كائن القيمة `ScopeRef`:**
   * **منع الحالات غير المتناسقة (Preventing Inconsistent States):** في السابق كان `Scope` و `TargetID` حقلين منفصلين، مما قد يؤدي لبيانات مشوهة (مثل: نطاق عام `GLOBAL` مع وجود `TargetID` عشوائي، أو نطاق مشروع بدون `TargetID`).
   * **النمذجة ككائن قيمة (Value Object Pattern):** النطاق في الحقيقة ليس مجرد مستوى، بل هو زوج متكامل `(Level, TargetID)`. دمجهما في `ScopeRef` يجعل النطاق كائناً غير قابل للتجزئة، ويوفر دوال مساعدة ضامنة لصحة البيانات مثل `GlobalScope()` و `ScopedTarget(level, id)`.
   * **تبسيط الواجهات وتوقيع الدوال:** بدلاً من تمرير معاملين في كل دالة `(scope perms.ScopeLevel, targetID string)`، يتم تمرير كائن واحد متماسك `scope ScopeRef`.

### ج. واجهات التكامل الخدمي (Service Provider Interfaces - SPI)

1. **مزود تسلسل الهويات (`PrincipalHierarchyProvider`)**:

   ```go
   // PrincipalHierarchyProvider يوفر قائمة الهويات الموروثة/المرتبطة بالهوية المستعلم عنها
   // بناءً على نوع الفاعل ومعرفه (مثل: USER ومُعرّفه -> [معرف المستخدم، مجموعاته المباشرة، مجموعة كل المستخدمين])
   type PrincipalHierarchyProvider interface {
       ResolveIdentities(principalType PrincipalType, principalID string) ([]string, error)
   }
   ```

2. **مزود تسلسل النطاقات (`ScopeHierarchyProvider`)**:

   ```go
   // ScopeHierarchyProvider يوفر معلومات العلاقات بين النطاقات المختلفة
   // (مثل: نطاق المشروع -> نطاق المنظمة الأب -> النطاق العام)
   type ScopeHierarchyProvider interface {
       GetParentScopes(scope ScopeRef) ([]ScopeRef, error)
       ValidateScope(scope ScopeRef) error
   }
   ```

### د. واجهة خدمة الأدوار المجردة (`IRolesService`)

```go
type IRolesService interface {
    // إدارة الأدوار (Role Lifecycle)
    GetRole(id string) (Role, error)
    GetAllRoles() []Role
    CreateRole(actorPerms []string, role Role) (*Role, error)
    UpdateRole(actorPerms []string, id string, name string, description string, permissions []string) (*Role, error)
    DeleteRole(actorPerms []string, id string) error
    CloneRole(actorPerms []string, sourceRoleID string, newRoleID string, newName string, newDescription string) (*Role, error)
    MergeRoles(actorPerms []string, targetRoleID string, sourceRoleIDs []string) error

    // إدارة التعيينات المجردة (Assignment Management)
    AssignRole(actorPerms []string, roleID string, principalType PrincipalType, principalID string, scope ScopeRef) (*RoleAssignment, error)
    RevokeAssignment(actorPerms []string, assignmentID string) error
    GetAssignmentsForPrincipal(principalType PrincipalType, principalID string) []RoleAssignment
    GetAssignmentsForScope(scope ScopeRef) []RoleAssignment

    // حساب الصلاحيات المجرد (Effective Permissions Evaluation)
    GetEffectivePermissions(principalType PrincipalType, principalID string, scope ScopeRef) ([]string, error)
    HasPermission(principalType PrincipalType, principalID string, scope ScopeRef, requiredPerm string) (bool, error)
}
```

---

## 5. مراحل الخطة التنفيذية خطوة بخطوة (Execution Phases)

### المرحلة الأولى: تأسيس طبقة التجريد والواجهات (SPI)

1. إنشاء ملف `internal/services/roles_service/spi.go` يحتوي على:
   * تعريف `ScopeRef`.
   * واجهة `PrincipalHierarchyProvider`.
   * واجهة `ScopeHierarchyProvider`.
   * توفير تطبيق افتراضي في الذاكرة (`InMemoryProvider`) لسهولة الاختبار والاستخدام المستقل.

### المرحلة الثانية: تطهير حزمة `roles_service` من الكيانات الخارجية

1. تعديل `assignment.go`:
   * حذف `struct Group` و `struct Project` و `struct Organization`.
   * تحديث `RoleAssignment` ليستخدم `ScopeRef` مع الاحتفاظ بـ `PrincipalType` و `PrincipalID`.
   * إزالة أخطاء الكيانات الخارجية المحددة واستبدالها بأخطاء تجريدية (`ErrTargetNotFound`, `ErrInvalidScope`).
2. تعديل `service.go`:
   * إزالة حقول `organizations`, `projects`, `groups`, `userGroups` من `struct Service`.
   * إزالة دوال التسجيل والإدارة الخاصة بالكيانات الملموسة.
   * حقن مزودي `PrincipalHierarchyProvider` و `ScopeHierarchyProvider` في باني الخدمة `NewService(...)`.

### المرحلة الثالثة: إعادة بناء محرك تجميع الصلاحيات (`aggregation.go`)

1. استبدال خوارزمية البحث المباشر في المجموعات باستدعاء `PrincipalHierarchyProvider.ResolveIdentities(principalType, principalID)`.
2. استبدال كود ربط المشروع بالمنظمة باستدعاء `ScopeHierarchyProvider.GetParentScopes(scope)`.
3. إزالة كود الفرق الخاصة بالمشروع (`TeamMemberIDs`, `TeamRoleID`) ليكون التعامل معها عبر تعيينات قياسية أو عبر مزود الهويات والنطاقات.
4. الإبقاء الصارم على قواعد يوتراك لتجاهل النطاقات (`Disregard Rules`) وفق معايير النطاق المجرد.

### المرحلة الرابعة: إنشاء خدمة السياسات المنفصلة (`policy_service`)

1. إنشاء حزمة جديدة `internal/services/policy_service`:
   * نقل منطق `visibility.go` إليها بالكامل.
   * نقل `IssueContext` و `ArticleContext`.
   * توفير دوال:
     * `CanViewIssue(userID string, issue IssueContext)`
     * `CanViewArticle(userID string, article ArticleContext)`
     * `CanAccessIssuePrivateField(userID string, projectID string, isUpdate bool)`
     * `CheckInherentAccess(...)`
   * ربط هذه الخدمة بـ `roles.IRolesService` و `perms.IPermissionService`.

### المرحلة الخامسة: الاختبار والتحقق المتكامل (Testing & Validation)

1. كتابة اختبارات وحدة مستقلة لخدمة الأدوار المجردة `service_test.go` باستخدام مزودات وهمية (Mock Providers).
2. كتابة اختبارات خدمة السياسات `policy_service_test.go` للتحقق من استمرار عمل كافة سيناريوهات YouTrack المعتمدة على التذاكر والمقالات.
3. تشغيل `go test ./...` وضمان نجاح جميع الاختبارات مع تغطية كاملة لكافة الحالات.

---

## 6. مصفوفة المقارنة المعمارية (Before vs. After)

| وجه المقارنة | الوضع الحالي (Coupled) | الوضع المستهدف بعد التجريد (Decoupled) |
| :--- | :--- | :--- |
| **طبيعة الخدمة** | تطبيق مونوثيلي يحتوي على بيانات المشاريع والتذاكر والمجموعات | محرك RBAC نقي ومجرد تماماً |
| **الاعتماد على الكيانات** | مرتبطة بـ `Group`, `Project`, `Organization`, `Issue`, `Article` | **صفر ارتباط**؛ تعتمد فقط على معرفات عامة `PrincipalID` و `ScopeRef` |
| **هرمية الهويات** | جداول داخلية مشفرة في الخدمة (`userGroups`) | واجهة خارجية مرنة: `PrincipalHierarchyProvider` |
| **هرمية النطاقات** | حقل ثابت في نموذج المشروع (`OrganizationID`) | واجهة خارجية مرنة: `ScopeHierarchyProvider` |
| **تقييم صلاحيات التذاكر** | داخل خدمة الأدوار (`roles.Service`) | في خدمة سياسات مخصصة (`policy_service`) |
| **إعادة الاستخدام** | صعبة ومحصورة فقط بتطبيق شبيه بيوتراك القديم | يمكن استخدامها في أي تطبيق أو نظام صلاحيات متعدد النطاقات |

---

## 7. خطة الاختبار والتحقق من الجودة (Testing & QA Plan)

1. **اختبارات عزل خدمة الأدوار (Roles Isolation Tests)**:
   * التأكد من إنشاء وتحديث ودمج واستنساخ الأدوار دون وجود أي مشاريع أو مجموعات مسجلة.
   * التحقق من سلامة الحماية ضد تصعيد الصلاحيات (`checkEscalation`).
   * التحقق من حماية قراءة الأدوار المدمجة (`IsReadOnly`).
2. **اختبارات محرك التجميع عبر المزودات (Provider Aggregation Tests)**:
   * محاكاة هويات متعددة (مستخدم + 3 مجموعات متداخلة) والتحقق من اتحاد الصلاحيات (`Union`).
   * محاكاة نطاقات متعددة المستويات والتحقق من تطبيق قواعد التجاهل (`Disregard Rules`).
3. **اختبارات خدمة السياسات المنسوخة (Policy Tests)**:
   * استمرار اجتياز سيناريوهات رؤية التذاكر (عامة، مقيدة، صلاحية المسؤول، وصاحب التذكرة).
   * استمرار اجتياز وراثة المقالات المعرفية الشجرية (`Article Hierarchy`).
   * التأكد من الحقول الخاصة بالتذاكر والحقوق المتأصلة (`Inherent Rights`).
4. **فحص الامتثال والأداء**:
   * تشغيل `go vet ./...` و `go test -race ./...`.
