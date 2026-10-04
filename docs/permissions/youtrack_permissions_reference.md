# المرجع الشامل لتحليل صلاحيات يوتراك (YouTrack Permissions Reference & Analysis)

**المصدر الرسمي:** [JetBrains YouTrack Server Documentation - Permissions Reference](https://www.jetbrains.com/help/youtrack/server/youtrack-permissions-reference.html)  
**إصدار النظام المُحلل:** YouTrack Server 2026.1 / 2026.2  
**تاريخ التحليل والتضمين:** أكتوبر 2026  

---

## 1. نظرة عامة على نموذج الصلاحيات في YouTrack

يعتمد نظام YouTrack في إدارة الوصول على نموذج **التحكم بالوصول القائم على الأدوار (Role-Based Access Control - RBAC)**:

- **الصلاحية (Permission):** هي تفويض محدد يُمنح للمستخدم لتنفيذ إجراء أو عملية معينة على كيان أو مورد داخل النظام. لا تُمنح الصلاحيات للمستخدمين مباشرة، بل تُجمع داخل "أدوار" (Roles).
- **الدور (Role):** هو مجموعة محددة من الصلاحيات تُعين لمستخدم أو مجموعة مستخدمين ضمن نطاق سياقي معين (Global، Organization، أو Project).
- **النطاق السياقي (Context Scope):** يحدد الحاوية أو الحدود التي تكون فيها الصلاحيات الممنوحة للمستخدم فعالة ونافذة.

---

## 2. مستويات نطاقات الصلاحيات (Permission Scopes)

تُصنف الصلاحيات في YouTrack بحسب النطاق الجغرافي/التنظيمي الذي تؤثر فيه إلى ثلاثة مستويات:

### 1. الصلاحيات العامة (Global Permissions)

- تُمنح على مستوى النظام بالكامل وتكون مستقلة عن أي مشروع محدد.

- من أمثلتها: إنشاء المستخدمين (`CREATE_USER`)، إنشاء المؤسسات (`CREATE_ORGANIZATION`)، وإنشاء المشاريع الجديدة (`CREATE_PROJECT`).
- لا يمكن تقييد هذه الصلاحيات بمشروع أو مؤسسة بعينها، وتظهر في واجهة YouTrack بوسم **Global**.

### 2. صلاحيات المؤسسة (Organization-Scoped Permissions)

- تنطبق الصلاحيات ضمن حدود مؤسسة تنظيمية معينة وتؤثر على جميع المشاريع التابعة لتلك المؤسسة.

- من أمثلتها: تعديل بيانات المؤسسة وتعيين المشاريع (`UPDATE_ORGANIZATION`)، وقراءة بيانات المؤسسة (`READ_ORGANIZATION`).

### 3. صلاحيات المشروع (Project-Scoped Permissions)

- تنطبق فقط على مشروع بعينه أو مجموعة مشاريع محددة يُمنح المستخدم دوراً فيها.

- من أمثلتها: قراءة التذاكر (`READ_ISSUE`)، إنشاء تذكرة (`CREATE_ISSUE`)، وتعديل الحقول الخاصة (`PRIVATE_UPDATE_ISSUE`).
- إذا امتلك المستخدم دوراً يمنحه `READ_ISSUE` في "المشروع أ"، فإنه لا يمتلك أي وصول لتذاكر "المشروع ب" ما لم يُمنح دوراً فيه.

### قواعد عزل وانتشار النطاقات (Scope Isolation & Propagation Rules)

في إصدار 2026 تم وضع قواعد واضحة عند تعيين الأدوار ذات الصلاحيات المختلطة:

1. **عند تعيين دور على المستوى العام (Global Scope):** تصبح جميع صلاحيات الدور سارية ومفعلة.
2. **عند تعيين دور على مستوى المؤسسة (Organization Scope):** الصلاحيات العامة **لا تنتشر ولا تسري**، بينما تسري صلاحيات المؤسسة والمشاريع التابعة لها.
3. **عند تعيين دور على مستوى المشروع (Project Scope):** الصلاحيات العامة وصلاحيات المؤسسة **لا تسري ولا يكون لها أي أثر** داخل المشروع.

---

## 3. الصلاحيات المتأصلة للكتّاب ومقدمي البلاغات (Inherent Permissions)

يتبنى YouTrack مبدأ الأذونات المتأصلة (Inherent / Implied Rights)، حيث يحصل منشئ العنصر تلقائياً على حقوق قراءة وتعديل موارده الذاتية دون اشتراط امتلاك صلاحيات عامة واسعة:

| الكيان / الحالة | الصلاحية الممنوحة | الحقوق المتأصلة تلقائياً (Inherent Rights) | ما لا تشمله الصلاحية المتأصلة |
| :--- | :--- | :--- | :--- |
| **مقدم البلاغ (Issue Reporter)** | `CREATE_ISSUE` | - عرض الحقول العامة للتذكرة التي أنشأها.<br>- تعديل الحقول العامة لتذكرته.<br>- إضافة روابط تذاكر لتذكرته.<br>(حتى بدون امتلاك `READ_ISSUE` أو `UPDATE_ISSUE` أو `LINK_ISSUE`). | **حذف التذكرة:** لا يمكن للمستخدم حذف تذكرته الذاتية دون امتلاك صلاحية صريحة لـ `DELETE_ISSUE`. |
| **رافع المرفقات (File Attacher)** | `CREATE_ATTACHMENT_ISSUE` | - تعديل الملفات المرفقة التي رفعها.<br>- تقييد ظهورها (Visibility Restriction) دون الحاجة لصلاحية `UPDATE_ATTACHMENT_ISSUE`. | تعديل مرفقات المستخدمين الآخرين. |
| **حذف المرفقات الذاتية** | جميع المستخدمين | - يحق لأي مستخدم حذف الملفات التي أرفقها بنفسه دون الحاجة لصلاحية `DELETE_ATTACHMENT_ISSUE`. | حذف ملفات رفعها غيره. |
| **كاتب تعليق التذكرة** | `CREATE_COMMENT` | - قراءة التعليقات التي كتبها بنفسه دون امتلاك `READ_COMMENT`. | تعديل تعليقه يتطلب صلاحية `UPDATE_COMMENT`. |
| **مسجل بنود العمل (Work Item Author)** | `CREATE_WORK_ITEM` | - قراءة بنود العمل (سجلات الوقت) الخاصة به دون امتلاك `READ_WORK_ITEM`. | تعديل بند العمل يتطلب `UPDATE_WORK_ITEM`. |
| **كاتب تعليق المقال (Knowledge Base)** | `CREATE_ARTICLE_COMMENT` | - قراءة وتعديل تعليقاته الذاتية على مقالات قاعدة المعرفة دون الحاجة لصلاحية `UPDATE_ARTICLE_COMMENT`. | تعديل أو حذف تعليقات المستخدمين الآخرين. |

> 📖 **للاطلاع على الشرح المعمق لكل إجراء وثوابت `InherentAction`:** راجع الدليل المخصص [inherent_permissions_guide.md](./inherent_permissions_guide.md).

---

## 4. الصلاحيات الضمنية والتابعة (Implied and Dependent Permissions)

لتسهيل إعداد الأدوار وتجنب الأخطاء التشغيلية، يربط YouTrack الصلاحيات عبر شبكة علاقات تكاملية:

- **الصلاحية الضمنية (Implied Permission):** عند إضافة صلاحية رئيسية لدور ما، تُضاف الصلاحيات الضمنية التي تعتمد عليها برمجياً وتقنياً إلى الدور **تلقائياً**.
  - *مثال:* لا يمكن إنشاء تذكرة دون معرفة اسم ومعرف المشروع؛ لذلك صلاحية `CREATE_ISSUE` تتضمن ضمنياً (`Implies`) صلاحية `READ_PROJECT_BASIC`.
- **الصلاحية التابعة (Dependent Permission):** عند إزالة أو سحب صلاحية أساسية من دور، تُحذف وتُلغى تلقائياً جميع الصلاحيات التي تعتمد عليها (`Cascading Revocation`).
  - *مثال:* إذا أزيلت صلاحية `READ_PROJECT_BASIC` من دور، تُحذف تلقائياً صلاحيات `CREATE_ISSUE` و `READ_ISSUE` و `UPDATE_PROJECT`.

> 📊 **المخططات الشجرية للعلاقات:**
>
> - [youtrack_permissions_tree.mmd](./diagrams/youtrack_permissions_tree.mmd): مخطط مصنف حسب حاويات الوحدات.
> - [youtrack_permissions_hierarchy.mmd](./diagrams/youtrack_permissions_hierarchy.mmd): مخطط شجري نقي خالص وممتد مباشرة من العقدة الجذرية بدون حاويات (Flat Tree).

---

## 5. التحديثات الجوهرية لنظام الصلاحيات في إصدارات 2026 (2026.1 / 2026.2 Updates)

شهدت إصدارات YouTrack 2026 تبسيطاً وتحديثاً جذرياً لبنية الصلاحيات:

1. **التقارير كميزة اختيارية (Reports as an Optional Feature):**
   - تم إلغاء صلاحيات التقارير المستقلة السابقة (`Create Report`, `Read Report`, `Share Report`).
   - أصبحت إدارة التقارير ميزة اختيارية على مستوى النظام تُفعل لمجموعات مستخدمين محددة.
2. **إعادة هيكلة صلاحيات المجموعات (Groups Scheme Modernization):**
   - تم استبدال صلاحية `Read Group` بميزة اختيارية (`Read Groups feature`).
   - مستخدمو `Update Project` و `Low-level Admin Read` يمتلكون تلقائياً صلاحية قراءة المجموعات.
   - إنشاء وحذف المجموعات متاح لمستخدمي `Update Project` أو `Low-level Admin Write`.
3. **دمج إدارة الأدوار (Role Management Consolidation):**
   - أُلغيت صلاحيات الأدوار المستقلة السابقة (`Read Role`, `Manage Role`).
   - أصبحت إدارة وتعيين الأدوار متأصلة ضمن صلاحية `UPDATE_PROJECT` على مستوى المشروع، وضمن `UPDATE_ORGANIZATION` على مستوى المؤسسة، وضمن `ADMIN_UPDATE_APP` على المستوى العام.
4. **استحداث صلاحيات محتوى التطبيقات وسير العمل (App & Workflow Content Permissions):**
   - أُضيفت صلاحيتا `READ_APP_CONTENT` و `UPDATE_APP_CONTENT` للتحكم في البرمجيات النصية والنصوص البرمجية وملفات التكوين الخاصة بالتطبيقات وقواعد سير العمل وسياسات SLA على مستوى المشروع.

---

## 6. الفهرس الكامل لجميع صلاحيات YouTrack (Complete Permissions Reference)

تحتوي القائمة التالية على جميع الصلاحيات المعرفة في YouTrack مصنفة حسب الوحدات الوظيفية:

| المعرف البرمجي (Key) | الاسم المعروض (Display Name) | الكيان (Entity) | النطاق (Scope) | العملية | الصلاحيات المتضمنة (Implied) | الصلاحيات التابعة (Dependent) | الوصف المختصر |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `ADMIN_READ_APP` | Low-level Admin Read | SYSTEM | GLOBAL | READ | - | `ADMIN_UPDATE_APP` | قراءة الإعدادات الإدارية المنخفضة والقياسات والمجموعات والأدوار. |
| `ADMIN_UPDATE_APP` | Low-level Admin Write | SYSTEM | GLOBAL | ADMIN | `ADMIN_READ_APP` | - | إدارة الإجراءات الإدارية المتقدمة والنسخ الاحتياطي وإدارة المجموعات والأدوار عالمياً. |
| `READ_APP_CONTENT` | Read App Content | APP | PROJECT | READ | - | `UPDATE_APP_CONTENT` | قراءة محتوى سكريبتات التطبيقات وسير العمل وسياسات SLA وسجلاتها. |
| `UPDATE_APP_CONTENT` | Update App Content | APP | PROJECT | UPDATE | `READ_APP_CONTENT` | - | إنشاء وتعديل واستيراد وحذف سكريبتات التطبيقات وسير العمل وتصحيحها برمجياً. |
| `CREATE_ARTICLE` | Create Article | ARTICLE | PROJECT | CREATE | `READ_ARTICLE` | - | إنشاء مقالات جديدة في قاعدة المعرفة للمشروع. |
| `DELETE_ARTICLE` | Delete Article | ARTICLE | PROJECT | DELETE | `READ_ARTICLE` | - | حذف المقالات من قاعدة المعرفة للمشروع. |
| `READ_ARTICLE` | Read Article | ARTICLE | PROJECT | READ | - | `CREATE_ARTICLE`, `DELETE_ARTICLE`, `UPDATE_ARTICLE` | قراءة المقالات ومحتواها داخل قاعدة المعرفة للمشروع. |
| `UPDATE_ARTICLE` | Update Article | ARTICLE | PROJECT | UPDATE | `READ_ARTICLE` | - | تعديل المقالات القائمة في قاعدة المعرفة للمشروع. |
| `CREATE_ARTICLE_COMMENT` | Create Article Comment | ARTICLE_COMMENT | PROJECT | CREATE | `READ_ARTICLE_COMMENT` | - | إضافة تعليقات على مقالات قاعدة المعرفة (ويتضمن تعديل وحذف تعليقات الكاتب الذاتية). |
| `DELETE_ARTICLE_COMMENT` | Delete Article Comment | ARTICLE_COMMENT | PROJECT | DELETE | `READ_ARTICLE_COMMENT` | - | حذف أي تعليق على مقالات المشروع بما في ذلك تعليقات الآخرين. |
| `READ_ARTICLE_COMMENT` | Read Article Comment | ARTICLE_COMMENT | PROJECT | READ | - | `CREATE_ARTICLE_COMMENT`, `DELETE_ARTICLE_COMMENT`, `UPDATE_ARTICLE_COMMENT` | عرض التعليقات المنشورة على مقالات المشروع. |
| `UPDATE_ARTICLE_COMMENT` | Update Article Comment | ARTICLE_COMMENT | PROJECT | UPDATE | `READ_ARTICLE_COMMENT` | - | تعديل التعليقات المنشورة على المقالات بما فيها تعليقات الآخرين. |
| `APPLY_COMMANDS_SILENTLY` | Apply Commands Silently | ISSUE | PROJECT | SPECIAL | - | - | تطبيق الأوامر على التذاكر دون إرسال إشعارات للمشتركين. |
| `CREATE_ISSUE` | Create Issue | ISSUE | PROJECT | CREATE | `READ_PROJECT_BASIC` | - | إنشاء تذاكر جديدة في المشروع، مع التمتع بحقوق قراءة وتعديل الحقول العامة والربط لتذاكر الكاتب. |
| `DELETE_ISSUE` | Delete Issue | ISSUE | PROJECT | DELETE | - | - | حذف التذاكر نهائياً من المشروع. |
| `LINK_ISSUE` | Link Issues | ISSUE | PROJECT | LINK | - | - | ربط التذاكر بروابط علاقات (مكررة، تابعة، متعلقة). |
| `READ_HIDDEN_STUFF` | Override Visibility Restrictions | ISSUE | PROJECT | SPECIAL | `PRIVATE_READ_ISSUE` | - | تجاوز قيود الرؤية ومشاهدة التذاكر والتعليقات والمرفقات المخفية. |
| `READ_ISSUE` | Read Issue | ISSUE | PROJECT | READ | `READ_PROJECT_BASIC` | - | قراءة التذاكر وحقولها العامة. |
| `PRIVATE_READ_ISSUE` | Read Issue Private Fields | ISSUE | PROJECT | READ | `READ_PROJECT_BASIC` | `READ_HIDDEN_STUFF`, `PRIVATE_UPDATE_ISSUE` | قراءة الحقول الخاصة والمقيدة في التذاكر. |
| `UPDATE_ISSUE` | Update Issue | ISSUE | PROJECT | UPDATE | - | `PRIVATE_UPDATE_ISSUE` | تعديل قيم الحقول العامة للتذاكر. |
| `PRIVATE_UPDATE_ISSUE` | Update Issue Private Fields | ISSUE | PROJECT | UPDATE | `PRIVATE_READ_ISSUE`, `UPDATE_ISSUE` | - | تعديل قيم الحقول الخاصة والمحمية في التذاكر. |
| `UPDATE_WATCHERS` | Update Watchers | ISSUE | PROJECT | UPDATE | - | - | إضافة وإزالة مستخدمين آخرين من قائمة متابعي التذكرة. |
| `VIEW_VOTERS` | View Voters | ISSUE | PROJECT | READ | `READ_PROJECT_BASIC` | - | مشاهدة قائمة المستخدمين الذين صوتوا للتذكرة. |
| `VIEW_WATCHERS` | View Watchers | ISSUE | PROJECT | READ | `READ_PROJECT_BASIC` | - | مشاهدة قائمة المستخدمين المتابعين للتذكرة. |
| `CREATE_ATTACHMENT_ISSUE` | Add Attachment | ISSUE_ATTACHMENT | PROJECT | CREATE | - | - | إرفاق ملفات بالتذاكر (مع حق الكاتب بتعديلها وتقييد رؤيتها). |
| `DELETE_ATTACHMENT_ISSUE` | Delete Attachment | ISSUE_ATTACHMENT | PROJECT | DELETE | - | - | حذف أي ملف مرفق بالتذكرة بما في ذلك ملفات الآخرين. |
| `UPDATE_ATTACHMENT_ISSUE` | Update Attachment | ISSUE_ATTACHMENT | PROJECT | UPDATE | - | - | تعديل المرفقات وتغيير إعدادات ظهورها. |
| `CREATE_COMMENT` | Create Issue Comment | ISSUE_COMMENT | PROJECT | CREATE | - | - | إضافة تعليقات على التذاكر (ويتضمن حق الكاتب بقراءة تعليقاته). |
| `DELETE_COMMENT` | Delete Issue Comment | ISSUE_COMMENT | PROJECT | DELETE | - | - | حذف التعليقات من التذاكر. |
| `DELETE_NOT_OWN_COMMENT` | Delete Not Own & Permanent Delete | ISSUE_COMMENT | PROJECT | DELETE | `READ_COMMENT` | - | حذف تعليقات المستخدمين الآخرين وحذف التعليقات نهائياً. |
| `READ_COMMENT` | Read Issue Comment | ISSUE_COMMENT | PROJECT | READ | - | `DELETE_NOT_OWN_COMMENT`, `UPDATE_NOT_OWN_COMMENT` | قراءة التعليقات المكتوبة على التذاكر. |
| `UPDATE_COMMENT` | Update Issue Comment | ISSUE_COMMENT | PROJECT | UPDATE | - | - | تعديل التعليقات المضافة للتذاكر. |
| `UPDATE_NOT_OWN_COMMENT` | Update Not Own Issue Comment | ISSUE_COMMENT | PROJECT | UPDATE | `READ_COMMENT` | - | تعديل تعليقات المستخدمين الآخرين على التذاكر. |
| `CREATE_NOT_OWN_WORK_ITEM` | Create Not Own Work Item | ISSUE_WORK_ITEM | PROJECT | CREATE | `CREATE_WORK_ITEM` | - | تسجيل ساعات عمل نيابة عن مستخدمين آخرين. |
| `CREATE_WORK_ITEM` | Create Work Item | ISSUE_WORK_ITEM | PROJECT | CREATE | - | `CREATE_NOT_OWN_WORK_ITEM` | إضافة سجلات وقت وبنود عمل للتذاكر (مع حق قراءة بنوده الذاتية). |
| `READ_WORK_ITEM` | Read Work Item | ISSUE_WORK_ITEM | PROJECT | READ | - | `UPDATE_NOT_OWN_WORK_ITEM` | قراءة قائمة بنود العمل وسجلات تتبع الوقت في التذكرة. |
| `UPDATE_NOT_OWN_WORK_ITEM` | Update Not Own Work Item | ISSUE_WORK_ITEM | PROJECT | UPDATE | `READ_WORK_ITEM`, `UPDATE_WORK_ITEM` | - | تعديل بنود عمل وسجلات وقت سجلها مستخدمون آخرون. |
| `UPDATE_WORK_ITEM` | Update Work Item | ISSUE_WORK_ITEM | PROJECT | UPDATE | - | `UPDATE_NOT_OWN_WORK_ITEM` | تعديل بنود العمل الذاتية المسجلة على التذاكر. |
| `CREATE_ORGANIZATION` | Create Organization | ORGANIZATION | GLOBAL | CREATE | `READ_ORGANIZATION` | - | إنشاء مؤسسات جديدة في النظام. |
| `DELETE_ORGANIZATION` | Delete Organization | ORGANIZATION | ORGANIZATION | DELETE | `READ_ORGANIZATION` | - | حذف المؤسسات وسجلاتها نهائياً من النظام. |
| `READ_ORGANIZATION` | Read Organization | ORGANIZATION | ORGANIZATION | READ | - | `CREATE_ORGANIZATION`, `DELETE_ORGANIZATION`, `UPDATE_ORGANIZATION` | قراءة بيانات وخصائص المؤسسة والمشاريع والأدوار المسندة لها. |
| `UPDATE_ORGANIZATION` | Update Organization | ORGANIZATION | ORGANIZATION | UPDATE | `READ_ORGANIZATION` | - | تعديل خصائص المؤسسة وإدارة تعيين المشاريع والأدوار والصلاحيات. |
| `CREATE_PROJECT` | Create Project | PROJECT | GLOBAL | CREATE | - | - | إنشاء مشاريع جديدة في النظام. |
| `DELETE_PROJECT` | Delete Project | PROJECT | PROJECT | DELETE | `READ_PROJECT` | - | حذف المشاريع بالكامل. |
| `READ_PROJECT_BASIC` | Read Project Basic | PROJECT | PROJECT | READ | - | شبكة الصلاحيات التابعة (`READ_PROJECT`, `CREATE_ISSUE`, `READ_ISSUE`, ...) | قراءة الخصائص الأساسية للمشروع (الاسم، الوصف، المالك، والشعار). |
| `READ_PROJECT` | Read Project Full | PROJECT | PROJECT | READ | `READ_PROJECT_BASIC` | `DELETE_PROJECT`, `UPDATE_PROJECT` | قراءة كامل خصائص المشروع والأدوار وفريق العمل والصلاحيات. |
| `UPDATE_PROJECT` | Update Project | PROJECT | PROJECT | UPDATE | `READ_PROJECT` | - | إدارة وتعديل خصائص المشروع والمجموعات والأدوار والحقول وسير العمل والتطبيقات. |
| `CREATE_USER` | Create User | USER | GLOBAL | CREATE | - | - | إنشاء حسابات مستخدمين جديدة وإرسال دعوات التسجيل. |
| `DELETE_USER` | Delete User | USER | GLOBAL | DELETE | `READ_USER` | - | حذف حسابات المستخدمين من النظام. |
| `READ_USER_BASIC` | Read User Basic | USER | GLOBAL | READ | - | `READ_USER` | عرض قائمة المستخدمين وقراءة المعرف والاسم والصورة الرمزية. |
| `READ_USER` | Read User Details | USER | GLOBAL | READ | `READ_USER_BASIC` | `DELETE_USER`, `UPDATE_USER` | عرض تفاصيل الحساب الكاملة للمستخدمين المسجلين. |
| `UPDATE_PROFILE` | Update Self | USER | GLOBAL | UPDATE | - | `UPDATE_USER` | تعديل بيانات أمان الحساب الذاتي (كلمات المرور، التوثيق الثنائي، الرموز). |
| `UPDATE_USER` | Update User | USER | GLOBAL | UPDATE | `UPDATE_PROFILE`, `READ_USER` | - | تعديل ملفات المستخدمين وحظرهم ودمج حساباتهم وتجهيل بياناتهم الشخصية. |
| `CREATE_WATCH_FOLDER` | Create Tag or Saved Search | WATCH_FOLDER | PROJECT | CREATE | - | - | إنشاء الوسوم وعمليات البحث المحفوظة وطرق العرض المخصصة. |
| `DELETE_WATCH_FOLDER` | Delete Tag or Saved Search | WATCH_FOLDER | PROJECT | DELETE | - | - | حذف الوسوم وعمليات البحث وطرق العرض. |
| `UPDATE_WATCH_FOLDER` | Edit Tag or Saved Search | WATCH_FOLDER | PROJECT | UPDATE | - | - | تعديل الوسوم وعمليات البحث وطرق العرض المخصصة. |
| `SHARE_WATCH_FOLDER` | Share Custom View | WATCH_FOLDER | PROJECT | SHARE | - | - | مشاركة الوسوم وعمليات البحث مع مستخدمين أو مجموعات أخرى. |

---

## 7. تضمين التحليل البرمجي في خدمة Go (`internal/services/permissions_services`)

تم تضمين نتائج هذا التحليل بالكامل داخل حزمة Go الرسمية للخدمة:

1. **`permission.go`**:
   - تعريف الأنواع الأساسية (`ScopeLevel`, `EntityType`, `OperationType`).
   - هيكل `Permission` مع حقول التعاريف والتضمين والتبعية ووسوم JSON.
   - تعريف حالات الحقوق المتأصلة `InherentAction`.

2. **`catalog.go`**:
   - تعريف الثوابت النصية الرسمية لمعرفات الصلاحيات (`PermReadIssue`, `PermCreateIssue`, ...).
   - جدول الفهرس الكامل بجميع الصلاحيات.
   - دالة `BuildDefaultCatalog()` التي تبني السجل وتحسب الصلاحيات التابعة (`DependentPerms`) آلياً عبر عكس مخطط التضمين (`Graph Inversion`).

3. **`service.go`**:
   - واجهة `IPermissionService` وتنفيذها `Service`.
   - `ResolveImplied()`: حساب الإغلاق المتعدي (Transitive Closure) للصلاحيات الضمنية تلقائياً عند إضافة صلاحية لدور.
   - `ResolveRevocation()`: الإلغاء المتسلسل المتعدي (Cascading Revocation) لجميع الصلاحيات التابعة عند سحب صلاحية أساسية.
   - `ValidatePermissionsForScope()`: تطبيق قواعد عزل الصلاحيات حسب النطاق (Global vs Organization vs Project).
   - `CheckInherentAccess()`: التحقق التلقائي من حقوق منشئي التذاكر والمرفقات والتعليقات دون اشتراك أذونات موسعة.
   - `HasPermission()`: التحقق السريع من مطابقة الصلاحيات.

4. **`service_test.go`**:
   - اختبارات وحدة برمجية شاملة تضمن صحة السجل، تكامل علاقات التضمين والإلغاء، العزل النطاقي، والصلاحيات المتأصلة.
