# الدليل الشامل لمفاهيم الحقوق المتأصلة (Inherent Permissions) في YouTrack

يقدم هذا المستند توثيقاً وتفصيلاً شاملاً لنظام **الحقوق المتأصلة (Inherent Rights / Inherent Permissions)** وفق هندسة نظام إدارة الصلاحيات في JetBrains YouTrack، وتطبيقها البرمجي الممثل بالنوع `InherentAction` في هذا المشروع.

---

## 1. ما هي "الحقوق المتأصلة" (Inherent Permissions)؟

في أنظمة التحكم في الوصول التقليدية القائمة على الأدوار (RBAC)، تُمنح الصلاحيات عادةً بشكل عام على مستوى المشروع أو النظام ككل (Global/Project Scope). على سبيل المثال، إعطاء صلاحية "تعديل التذاكر `Update Issue`" يعني أن المستخدم يستطيع تعديل أي تذكرة في المشروع.

لكن YouTrack يطبق مبدأ أمنياً متقدماً يعرف بـ **الحقوق المتأصلة لملكية الموارد (Creator/Author Inherent Rights)**:
> **القاعدة الأساسية:** يحصل المستخدم الذي ينشئ مورداً معيناً (تذكرة، تعليق، مرفق، سجل عمل) تلقائياً على حقوق تشغيلية محددة على **ذلك المورد بعينه** دون الحاجة إلى منحه صلاحيات إدارية عامة وواسعة على مستوى المشروع بأكمله.

### الهدف المعماري من هذا المبدأ

1. **تطبيق مبدأ الحد الأدنى من الامتيازات (Least Privilege):** تمكين المستخدمين الخارجيين أو المبلّغين (Reporters) من متابعة بلاغاتهم دون فتح المشروع لهم.
2. **عزل الصلاحيات:** لا يحتاج المستخدم لصلاحية تعديل تذاكر الآخرين لكي يتمكن من تصحيح خطأ إملائي في تذكرته الخاصة.
3. **تبسيط تكوين الأدوار (Role Configuration):** يكفي منح الدور صلاحية إنشاء بسيطة (مثل `Create Issue`) وسيتولى محرك الصلاحيات منح الحقوق المتأصلة ذات الصلة تلقائياً.

---

## 2. التحليل التفصيلي لثوابت `InherentAction`

في الكود المصدري المطبّق في الملف [`permission.go`](../../internal/services/permissions_services/permission.go#L79-L94):

```go
type InherentAction string

const (
    InherentReadOwnIssuePublicFields   InherentAction = "READ_OWN_ISSUE_PUBLIC_FIELDS"
    InherentUpdateOwnIssuePublicFields InherentAction = "UPDATE_OWN_ISSUE_PUBLIC_FIELDS"
    InherentLinkOwnIssue               InherentAction = "LINK_OWN_ISSUE"
    InherentModifyOwnAttachment        InherentAction = "MODIFY_OWN_ATTACHMENT"
    InherentDeleteOwnAttachment        InherentAction = "DELETE_OWN_ATTACHMENT"
    InherentRestrictOwnAttachment      InherentAction = "RESTRICT_OWN_ATTACHMENT"
    InherentReadOwnIssueComment        InherentAction = "READ_OWN_ISSUE_COMMENT"
    InherentReadOwnWorkItem            InherentAction = "READ_OWN_WORK_ITEM"
    InherentReadOwnArticleComment      InherentAction = "READ_OWN_ARTICLE_COMMENT"
    InherentUpdateOwnArticleComment    InherentAction = "UPDATE_OWN_ARTICLE_COMMENT"
    InherentDeleteOwnArticleComment    InherentAction = "DELETE_OWN_ARTICLE_COMMENT"
)
```

فيما يلي شرح تفصيلي دقيق لكل مفهوم:

---

### أولاً: حقوق التذاكر الخاصة بمقدم البلاغ (Issue Reporter)

#### 1. `InherentReadOwnIssuePublicFields` (`READ_OWN_ISSUE_PUBLIC_FIELDS`)

- **المعنى:** حق قراءة وعرض الحقول العامة (Public Fields) للتذكرة التي أنشأها المستخدم بنفسه.
- **كيف يعمل؟** إذا كان المستخدم هو صاحب التذكرة (`isAuthorOrReporter = true`) ويمتلك صلاحية الإنشاء `CREATE_ISSUE`، يستطيع رؤية تفاصيل تذكرته الأساسية وحقولها العامة (كالعنوان، الوصف، الحالة العامة) حتى لو لم يكن دوره يمتلك صلاحية القراءة العامة للمشروع `READ_ISSUE`.
- **الاستثناء:** لا يمنحه هذا قراءة الحقول الخاصة أو الحساسة (Private Fields)؛ إذ تتطلب تلك صلاحية صريحة لـ `READ_ISSUE_PRIVATE_FIELDS`.

#### 2. `InherentUpdateOwnIssuePublicFields` (`UPDATE_OWN_ISSUE_PUBLIC_FIELDS`)

- **المعنى:** حق تعديل وتحديث الحقول العامة في التذكرة التي أنشأها المستخدم بنفسه.
- **كيف يعمل؟** يستطيع مقدم البلاغ تحديث وصف المشكلة، إضافة معلومات جديدة، أو تعديل البيانات العامة لتذكرته، طالما أنه صاحب التذكرة ويمتلك `CREATE_ISSUE`. لا يشترط هذا امتلاكه لصلاحية التعديل الشاملة `UPDATE_ISSUE` في المشروع.
- **الأمان:** يحافظ على سرية تذاكر الفريق الآخرين، بينما يمنح المبلّغ مرونة تعديل وتوضيح بلاغه الذاتي.

#### 3. `InherentLinkOwnIssue` (`LINK_OWN_ISSUE`)

- **المعنى:** حق إضافة وحذف الروابط بين تذكرته وتذاكر أخرى (مثل ربط التذكرة بكونها تابعة أو مكررة).
- **كيف يعمل؟** بمجرد امتلاك المستخدم لصلاحية `CREATE_ISSUE`، يكتسب حق ربط تذكرته دون الحاجة لمنحه صلاحية `LINK_ISSUE` على مستوى المشروع.

---

### ثانياً: حقوق المرفقات الخاصة برافع الملف (File Attacher)

#### 4. `InherentModifyOwnAttachment` (`MODIFY_OWN_ATTACHMENT`)

- **المعنى:** حق تعديل خصائص الملف المرفق الذي رفعه المستخدم (مثل تغيير اسم الملف أو وصفه).
- **كيف يعمل؟** إذا رفع المستخدم صورة أو مستنداً، يحق له تعديله متى ما امتلك صلاحية الرفع `CREATE_ATTACHMENT_ISSUE` (`PermAddAttachment`). لا يشترط امتلاك صلاحية `UPDATE_ATTACHMENT_ISSUE`.

#### 5. `InherentDeleteOwnAttachment` (`DELETE_OWN_ATTACHMENT`)

- **المعنى:** حق حذف الملف المرفق الذي قام المستخدم برفعه بنفسه.
- **ميزة جوهرية فريدة:** في نظام YouTrack وفي هذا التنفيذ البرمجي، **يحق لأي مستخدم حذف الملفات التي أرفقها بنفسه دون أي قيود أو اشتراط لصلاحية إضافية** (`DELETE_ATTACHMENT_ISSUE`).
- **السبب الأمني:** السماح للمستخدم بسحب أو حذف ملف حساس أو خاطئ قام برفعه بطريق الخطأ فوراً لحماية الخصوصية.

#### 6. `InherentRestrictOwnAttachment` (`RESTRICT_OWN_ATTACHMENT`)

- **المعنى:** حق تقييد مستوى رؤية الملف المرفق (Attachment Visibility Restriction).
- **كيف يعمل؟** يمكن لرافع الملف تحديد أن المرفق لا يراه إلا مجموعة معينة (مثلاً فريق التطوير فقط أو الإدارة) دون الحاجة لامتلاك صلاحية التعديل العامة للمرفقات `UPDATE_ATTACHMENT_ISSUE`.

---

### ثالثاً: حقوق التعليقات وسجلات العمل على التذاكر

#### 7. `InherentReadOwnIssueComment` (`READ_OWN_ISSUE_COMMENT`)

- **المعنى:** حق قراءة كاتب التعليق لتعليقاته الذاتية المنشورة على التذاكر.
- **كيف يعمل؟** إذا كتب المستخدم تعليقاً على تذكرة، يضمن النظام حقه في قراءة ما كتبه بنفسه طالما يمتلك صلاحية كتابة التعليق `CREATE_COMMENT` (`PermCreateIssueComment`)، حتى لو سُحبت منه صلاحية قراءة تعليقات المشروع العامة `READ_COMMENT`.

#### 8. `InherentReadOwnWorkItem` (`READ_OWN_WORK_ITEM`)

- **المعنى:** حق قراءة سجلات الوقت وبنود العمل (Time Tracking / Work Items) التي سجلها المستخدم بنفسه.
- **كيف يعمل؟** يحق للموظف/المستخدم الذي يسجل ساعات عمله على تذكرة استعراض سجلات أوقاته الذاتية طالما لديه `CREATE_WORK_ITEM` (`PermCreateWorkItem`)، دون الحاجة لصلاحية قراءة سجلات أوقات بقية الزملاء `READ_WORK_ITEM`.

---

### رابعاً: حقوق تعليقات المقالات في قاعدة المعرفة (Knowledge Base)

تختلف تعليقات مقالات قاعدة المعرفة (Articles) في YouTrack عن تعليقات التذاكر؛ إذ تتمتع باستقلالية أكبر للكاتب:

#### 9. `InherentReadOwnArticleComment` (`READ_OWN_ARTICLE_COMMENT`)

- **المعنى:** حق قراءة تعليقاته الذاتية على المقالات المنشورة في قاعدة المعرفة.
- **الشرط:** امتلاك صلاحية إنشاء التعليق `CREATE_ARTICLE_COMMENT`.

#### 10. `InherentUpdateOwnArticleComment` (`UPDATE_OWN_ARTICLE_COMMENT`)

- **المعنى:** حق تعديل وتصحيح نصوص تعليقاته الذاتية على مقالات المعرفة.
- **الشرط:** امتلاك صلاحية `CREATE_ARTICLE_COMMENT` فقط، دون الحاجة لصلاحية التعديل الإدارية `UPDATE_ARTICLE_COMMENT`.

#### 11. `InherentDeleteOwnArticleComment` (`DELETE_OWN_ARTICLE_COMMENT`)

- **المعنى:** حق حذف تعليقاته الذاتية على مقالات المعرفة.
- **الشرط:** امتلاك صلاحية `CREATE_ARTICLE_COMMENT`، فيتمكن من مسح تعليقه دون انتظار مدير النظام أو امتلاك `DELETE_ARTICLE_COMMENT`.

---

## 3. التحليل البرمجي: كيف تعمل دالة `CheckInherentAccess`؟

في الملف [`service.go`](../../internal/services/permissions_services/service.go#L242-L276)، يُقيّم النظام هذه الأفعال عبر الدالة التالية:

```go
func (s *Service) CheckInherentAccess(
    action InherentAction, 
    isAuthorOrReporter bool, 
    hasPermission func(string) bool,
) bool {
    // 1. صمام الأمان الأول: إذا لم يكن المستخدم هو الكاتب/المنشئ، تسقط الحقوق المتأصلة فوراً
    if !isAuthorOrReporter {
        return false
    }

    // 2. التحقق من طبيعة الفعل والشرط التمكيني المرتبط به
    switch action {
    case InherentReadOwnIssuePublicFields, InherentUpdateOwnIssuePublicFields, InherentLinkOwnIssue:
        return hasPermission(PermCreateIssue)

    case InherentModifyOwnAttachment, InherentRestrictOwnAttachment:
        return hasPermission(PermAddAttachment)

    case InherentDeleteOwnAttachment:
        // حق غير مشروط لأي مستخدم لحذف مرفقاته الذاتية
        return true

    case InherentReadOwnIssueComment:
        return hasPermission(PermCreateIssueComment)

    case InherentReadOwnWorkItem:
        return hasPermission(PermCreateWorkItem)

    case InherentReadOwnArticleComment, InherentUpdateOwnArticleComment, InherentDeleteOwnArticleComment:
        return hasPermission(PermCreateArticleComment)

    default:
        return false
    }
}
```

### منطق العمل الداخلي

1. **التحقق من الهوية والملكية (`isAuthorOrReporter`):**
   - إذا كان المورد يخص مستخدماً آخر، ترجع الدالة `false` فوراً، ويلزم حينئذ الرجوع إلى الصلاحيات الصريحة الشاملة (Explicit Role Permissions).
2. **الربط مع الصلاحية الأساسية (Base Permission Requirement):**
   - الحق المتأصل لا يُمنح في فراغ؛ فمثلاً لتعديل حقول تذكرتك، يجب أن تكون قادراً أصلاً على فتح تذكرة في المشروع (`PermCreateIssue`).
3. **الحالة الاستثنائية لـ `InherentDeleteOwnAttachment`:**
   - ترجع `true` دائماً؛ أي مستخدم يرفع مرفقاً يحق له حذفه دون اشتراط امتلاك أي صلاحية خاصة بحذف المرفقات.

---

## 4. أين ومتى يتم استدعاء دالة `CheckInherentAccess`؟

### أين يتم استدعاؤها معمارياً؟ (Where)

يتم استدعاء الدالة في **طبقة التحقق الأمني والتفويض (Authorization / Security Layer)** وتحديداً داخل خدمات النطاق (Domain Services) أو معالجات الطلبات (API Handlers/Controllers) عند تنفيذ عمليات محددة على مستوى مورد بعينه (Resource-Level Operations):

1. **في خدمة التذاكر (`IssueService` / `IssueHandler`):** عند طلب قراءة تذكرة، تحديث حقولها، أو ربطها بتذكرة أخرى.
2. **في خدمة المرفقات (`AttachmentService`):** عند حذف مرفق، تعديل بياناته، أو تقييد مستوى ظهوره.
3. **في خدمة التعليقات (`CommentService`):** عند عرض التعليقات أو تعديلها.
4. **في خدمة تتبع الوقت (`WorkItemService`):** عند استعراض ساعات العمل وبنوده.

### متى يتم استدعاؤها؟ (When)

تُستدعى الدالة في **المرحلة الثانية** من تسلسل اتخاذ القرار الأمني (Two-Tier Authorization Decision):

```text
               طلب المستخدم تنفيذ إجراء على مورد (Resource Action)
                                      │
                                      ▼
             ┌──────────────────────────────────────────────────┐
             │ فحص الصلاحية الصريحة العامة (Explicit Permission) │
             │  هل يمتلك المستخدم مثلاً: UPDATE_ISSUE للمشروع؟   │
             └────────────────────────┬─────────────────────────┘
                                      │
                         ┌────────────┴────────────┐
                         ▼ نعم                      ▼ لا
                 [ السماح بالإجراء ]       ┌─────────────────────────────────────┐
                                          │ استدعاء CheckInherentAccess(...)    │
                                          │  1. هل هو منشئ هذا المورد؟          │
                                          │  2. هل يملك صلاحية الإنشاء المطلوبة؟ │
                                          └──────────────────┬──────────────────┘
                                                             │
                                                ┌────────────┴────────────┐
                                                ▼ نعم                      ▼ لا
                                        [ السماح بالإجراء ]        [ رفض الطلب 403 ]
```

- **القاعدة الذهبية:** لا نلجأ إلى فحص الحق المتأصل إلا إذا كان المستخدم **لا يمتلك** الصلاحية العامة المباشرة؛ فإذا كان مدير المشروع يمتلك `UPDATE_ISSUE`، يُسمح له دون الحاجة للتأكد من كونه صاحب التذكرة.

### مثال كودي عملي (Go Implementation Pattern)

يوضح المثال التالي كيف يُستدعى `CheckInherentAccess` داخل معالج تحديث التذكرة (`UpdateIssueHandler`):

```go
func (h *IssueHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
    issueID := mux.Vars(r)["id"]
    currentUser := auth.GetUserFromContext(r.Context())
    issue := h.repo.FindByID(issueID)

    // دالة مساعدة لفحص صلاحيات المستخدم في مشروع هذه التذكرة
    hasPerm := func(permID string) bool {
        return h.permService.HasPermission(currentUser.ProjectPermissions[issue.ProjectID], permID)
    }

    // 1. هل يمتلك المستخدم صلاحية التعديل الشاملة في المشروع؟
    canUpdate := hasPerm(perms.PermUpdateIssue)

    // 2. إذا لم يملكها، نفحص الحق المتأصل لصاحب التذكرة:
    if !canUpdate {
        isReporter := (issue.ReporterID == currentUser.ID)
        // التحقق هل التعديل يقتصر على الحقول العامة وهل تنطبق عليه الصلاحية المتأصلة:
        canUpdate = h.permService.CheckInherentAccess(
            perms.InherentUpdateOwnIssuePublicFields, 
            isReporter, 
            hasPerm,
        )
    }

    if !canUpdate {
        http.Error(w, "ليس لديك صلاحية لتعديل هذه التذكرة", http.StatusForbidden)
        return
    }

    // المتابعة وتنفيذ التحديث...
}
```

---

## 5. جدول مقارنة شامل: الإجراءات المتأصلة والقيود الأمنية

| الإجراء المتأصل (`InherentAction`) | الكيان الهدف | الشرط التمكيني الأدنى | ما تشمله الصلاحية المتأصلة | ⛔ ما **لا** تشمله الصلاحية المتأصلة |
| :--- | :--- | :--- | :--- | :--- |
| `READ_OWN_ISSUE_PUBLIC_FIELDS` | التذاكر (Issues) | `CREATE_ISSUE` | قراءة الحقول العامة لتذكرته | قراءة الحقول الخاصة (`READ_ISSUE_PRIVATE_FIELDS`). |
| `UPDATE_OWN_ISSUE_PUBLIC_FIELDS` | التذاكر (Issues) | `CREATE_ISSUE` | تعديل العنوان، الوصف، الحقول العامة | تعديل تذاكر الآخرين أو الحقول الخاصة. |
| `LINK_OWN_ISSUE` | التذاكر (Issues) | `CREATE_ISSUE` | ربط تذكرته بتذاكر أخرى | تعديل روابط تذاكر الغير. |
| `MODIFY_OWN_ATTACHMENT` | المرفقات (Attachments) | `CREATE_ATTACHMENT_ISSUE` | إعادة تسمية المرفق وتعديل بياناته | تعديل مرفقات رفعها غيره. |
| `DELETE_OWN_ATTACHMENT` | المرفقات (Attachments) | لا شيء (تلقائي) | حذف المرفق الذي رفعه بنفسه | حذف مرفقات رفعها غيره. |
| `RESTRICT_OWN_ATTACHMENT` | المرفقات (Attachments) | `CREATE_ATTACHMENT_ISSUE` | ضبط وتحديد ظهور المرفق لفئات معينة | كسر التقييد المفروض على مرفقات الغير. |
| `READ_OWN_ISSUE_COMMENT` | تعليقات التذاكر | `CREATE_COMMENT` | قراءة تعليقاته التي كتبها | تعديل تعليق التذكرة (يتطلب صراحة `UPDATE_COMMENT`). |
| `READ_OWN_WORK_ITEM` | بنود العمل (Work Items) | `CREATE_WORK_ITEM` | قراءة سجلات أوقاته المسجلة | تعديل سجل الوقت (يتطلب صراحة `UPDATE_WORK_ITEM`). |
| `READ_OWN_ARTICLE_COMMENT` | تعليقات المقالات | `CREATE_ARTICLE_COMMENT` | قراءة تعليقاته على مقالات المعرفة | قراءة تعليقات مسودة مقال ليس له حق الوصول إليه. |
| `UPDATE_OWN_ARTICLE_COMMENT` | تعليقات المقالات | `CREATE_ARTICLE_COMMENT` | تعديل نصوص تعليقاته على المقال | تعديل تعليقات المستخدمين الآخرين. |
| `DELETE_OWN_ARTICLE_COMMENT` | تعليقات المقالات | `CREATE_ARTICLE_COMMENT` | حذف تعليقاته الذاتية على المقال | حذف تعليقات كتبها غيره. |

---

## 5. حدود أمنية هامة (Security Boundaries & Gotchas)

### 1. لا يوجد `InherentDeleteOwnIssue`

- **حذف التذاكر ليس حقاً متأصلاً:** لا يحق لمقدم البلاغ (Reporter) حذف التذكرة التي أنشأها لمجرد أنه منشئها.
- **السبب:** التذكرة قد تكون دخلت دورة عمل الفريق، رُبطت بها مهام فرعية، أو سجل عليها مهندسون أوقات عملهم. حذف التذكرة يتطلب دائماً صلاحية صريحة لـ `DELETE_ISSUE`.

### 2. الفارق بين تعليق التذكرة وتعليق المقال

- في **التذاكر (Issues)**: كاتب التعليق يملك حق قراءته (`READ_OWN_ISSUE_COMMENT`)، لكن تعديله أو حذفه يتطلب صلاحيات مخصصة (`UPDATE_COMMENT` / `UPDATE_OWN_COMMENT` و `DELETE_COMMENT` / `DELETE_OWN_COMMENT`).
- في **المقالات (Articles)**: منشئ التعليق يمتلك تلقائياً حق القراءة والتعديل والحذف (`READ`, `UPDATE`, `DELETE`) بمجرد امتلاكه لصلاحية إنشاء التعليق.

---

## 6. سيناريوهات عملية من واقع الاستخدام

### سيناريو 1: مستخدم خارجي يبلغ عن خطأ برمجي (Reporter Role)

- يملك فقط: `CREATE_ISSUE` في المشروع.
- **ما يستطيع فعله:**
  - فتح التذكرة.
  - استعراض الحقول العامة وتعديلها (مثل إضافة لقطة شاشة توضيحية أو تعديل الوصف).
  - ربط التذكرة برابط خارجي أو تذكرة أخرى.
  - حذف لقطة الشاشة التي رفعها إذا اكتشف أنها احتوت على بيانات شخصية بالخطأ.
- **ما لا يستطيع فعله:**
  - رؤية تذاكر العملاء الآخرين في المشروع (لأنه لا يملك `READ_ISSUE`).
  - حذف التذكرة بعد إنشائها.

### سيناريو 2: مستخدم يسجل ساعات عمله (Work Item Author)

- يملك فقط: `CREATE_WORK_ITEM`.
- يستطيع استعراض ساعات العمل التي قام بتسجيلها بنفسه للتأكد منها عبر `InherentReadOwnWorkItem` دون أن يطّلع على ساعات عمل زملائه في الفريق.

---

## 7. ملفات ذات صلة في المشروع

- التعريفات والأنواع: [`permission.go`](../../internal/services/permissions_services/permission.go#L79-L94)
- خدمة التحقق: [`service.go`](../../internal/services/permissions_services/service.go#L242-L276)
- حالات الاختبار: [`service_test.go`](../../internal/services/permissions_services/service_test.go#L163-L208)
- المرجع الشامل للصلاحيات: [`youtrack_permissions_reference.md`](./youtrack_permissions_reference.md#L53-L66)
