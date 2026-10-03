# تحليل وثيقة صلاحيات تطبيقات يوتراك (YouTrack App Permissions Analysis)

**رابط المصدر الرسمي:** [JetBrains YouTrack Developer Portal - App Permissions](https://www.jetbrains.com/help/youtrack/devportal/app-permissions.html)  
**سياق المنظومة:** بيئة تطوير تطبيقات وودجات يوتراك (YouTrack Apps & JavaScript Ecosystem)  
**تاريخ التوثيق:** أكتوبر 2026  

---

## 1. نظرة عامة: عن ماذا تتحدث هذه الصفحة؟

تُعد هذه الصفحة دليلاً مرجعياً رسمياً موجهاً لمطوري تطبيقات يوتراك (**YouTrack Apps / Widgets Developers**). تتناول الصفحة كيفية التحكم في **ظهور الودجات (Widget Visibility)** والتفاعل معها بناءً على نموذج الصلاحيات المعتمد في YouTrack، عبر تعريف حقل `permissions` في ملف بيان التطبيق (`manifest.json`).

توفر الصفحة إجابات دقيقة على أربعة أسئلة معمارية وأمنية:

1. **أين وكيف تُعرّف الصلاحيات برمجياً؟** داخل كائن الودجة (`widgets[]`) في ملف البيان كقائمة من المفاتيح النصية (`permissions: ["KEY_NAME"]`).
2. **ما هو المنطق الحسابي لتقييم الصلاحيات؟** تطبيق منطق **"أو" المنطقي (OR Logic / Disjunctive)**؛ إذ يكفي امتلاك المستخدم لواحدة فقط من الصلاحيات المحددة لفتح الودجة والتفاعل معها.
3. **كيف يُعالج سياق المشاريع (Project Scope)؟** التحقق يتم ديناميكياً لكل مشروع على حدة للمناطق الخاصة بالمشروع (مثل تفاصيل تذكرة معينة).
4. **ما هي قائمة المفاتيح الرسمية (Permission Keys) المعتمدة؟** جدول مرجعي شامل يربط بين الكيانات البرمجية (Entities) والعمليات ومفاتيحها المقابلة (`CREATE_ISSUE`، `READ_USER`، إلخ).

---

## 2. المبادئ المعمارية والأمنية الواردة في الصفحة

### 1. تقييد الرؤية القائم على الصلاحيات (Permission-Based Visibility Restrictions)

تتيح مصفوفة `permissions` لمدير التطبيق ومطوره إخفاء الودجات تماماً عن واجهة المستخدم إذا لم يكن لديه الحد الأدنى من الأذونات. هذا يمنع ازدحام الواجهة بعناصر لا يمكن للمستخدم الاستفادة منها، ويقلل من محاولات الوصول غير المصرح به.

```json
{
  "widgets": [
    {
      "key": "main-menu",
      "name": "Main Menu Item",
      "indexPath": "admin/index.html",
      "place": "MAIN_MENU_ITEM",
      "permissions": ["READ_USER"]
    }
  ]
}
```

### 2. قاعدة التحقق بالبدائل (The OR-Evaluation Rule)

تنص الوثيقة صراحة على:
> *"Users only need to be granted **one** of the required permissions to be able to view and interact with the widget, even when multiple permissions are specified in the manifest."*

- **الصيغة المنطقية:**  
  $$\text{CanAccessWidget} = \text{HasPerm}(P_1) \lor \text{HasPerm}(P_2) \lor \dots \lor \text{HasPerm}(P_n)$$
- **النتيجة المعمارية:** إذا وضعت `["READ_PROJECT", "CREATE_ISSUE"]`، فإن المستخدم الذي يملك فقط صلاحية إنشاء التذاكر سيتمكن من فتح الودجة حتى لو لم يكن يملك صلاحية قراءة المشروع كاملة.

### 3. الفحص الديناميكي لسياق المشروع (Per-Project Context Evaluation)

بالنسبة لنقاط التوسع التي تعمل داخل نطاق محدد لمشروع (Project-Scoped Extension Points مثل `ISSUE_BELOW_SUMMARY`):

- يتم تقييم الصلاحيات لكل مشروع على حدة.
- **مثال تطبيقي:** إذا تطلبت الودجة صلاحية ربط التذاكر `LINK_ISSUE` ووُضعت أسفل ملخص التذكرة، فإن المستخدم سيشاهد الودجة عند تصفحه تذاكر في **المشروع (أ)** حيث يمتلك صلاحية الربط، بينما تختفي نفس الودجة تلقائياً عند تصفح تذاكر **المشروع (ب)** إذا لم تُمنح له تلك الصلاحية فيه.

### 4. التحذير الأمني لطلبات REST وتأكيد العمليات (Security Caution)

تتضمن الوثيقة تنبيهاً أمنياً حاسماً بخصوص الودجات التفاعلية التي ترسل طلبات برمجية (REST Requests):
> *"If a widget sends REST requests that can change or delete data, restrict the widget to users who are expected to perform these actions. Request confirmation is stored for the current user only, so each user who can open the widget can confirm these requests on their own behalf."*

- **الخطر الأمني:** إذا احتوت الودجة على وظائف تعديل أو حذف بيانات عبر REST API، فإن إتاحتها للمستخدمين العاديين قد تمكّنهم من تأكيد تنفيذ طلبات لا يملكون كفاءة أو تفويضاً بإجرائها.
- **القاعدة الذهبية:** يجب ألا يقتصر تقييد الودجة على أدنى صلاحية قراءة، بل يجب ربط الودجة بصلاحيات التعديل/الحذف المناسبة للعمليات التي تنفذها خلف الكواليس.

---

## 3. مخطط سير عملية التحقق من ظهور الودجة (Widget Access Evaluation Flow)

```mermaid
flowchart TD
    Start([محاولة تحميل الودجة في واجهة المستخدم]) --> CheckPermsDef{هل تم تعريف حقل<br>permissions في manifest؟}
    
    CheckPermsDef -- لا يوجد قيد --> RenderWidget[عرض الودجة وتفعيلها للمستخدم]
    
    CheckPermsDef -- نعم يوجد --> CheckScope{هل نقطة التوسع<br>مرتبطة بنطاق مشروع معين؟<br>Project-Scoped Place}
    
    CheckScope -- نعم (مثل ISSUE_BELOW_SUMMARY) --> ContextEval[جلب صلاحيات المستخدم داخل سياق المشروع الحالي]
    CheckScope -- لا (مثل MAIN_MENU_ITEM) --> GlobalEval[جلب صلاحيات المستخدم الشاملة/العامة]
    
    ContextEval --> OrEvaluation{هل يمتلك المستخدم على الأقل<br>صلاحية واحدة من القائمة؟<br>(OR Logic)}
    GlobalEval --> OrEvaluation
    
    OrEvaluation -- نعم --> RenderWidget
    OrEvaluation -- لا --> HideWidget[إخفاء الودجة ومنع التفاعل معها]

    style RenderWidget fill:#e1f5fe,stroke:#03a9f4,stroke-width:2px;
    style HideWidget fill:#ffebee,stroke:#e91e63,stroke-width:2px;
```

---

## 4. الفهرس الشامل للكيانات ومفاتيح الصلاحيات الرسمية (Entity & Permission Keys)

توثق الصفحة **55 صلاحية رسمية** موزعة عبر **11 كياناً برمجياً (Entities)** أساسياً في YouTrack:

### 1. المقالات (Article - Knowledge Base)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Article | `CREATE_ARTICLE` | إنشاء مقالات جديدة في قاعدة المعرفة |
| Delete Article | `DELETE_ARTICLE` | حذف مقالات قاعدة المعرفة |
| Read Article | `READ_ARTICLE` | قراءة واستعراض مقالات قاعدة المعرفة |
| Update Article | `UPDATE_ARTICLE` | تعديل محتوى مقالات قاعدة المعرفة |

### 2. تعليقات المقالات (Article Comment)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Article Comment | `CREATE_ARTICLE_COMMENT` | إضافة تعليق على مقال |
| Delete Article Comment | `DELETE_ARTICLE_COMMENT` | حذف تعليق على مقال |
| Read Article Comment | `READ_ARTICLE_COMMENT` | قراءة التعليقات الواردة على المقالات |
| Update Article Comment | `UPDATE_ARTICLE_COMMENT` | تعديل تعليق على مقال |

### 3. التذاكر والمهام (Issue)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Apply Commands Silently | `APPLY_COMMANDS_SILENTLY` | تنفيذ أوامر التذاكر دون إرسال إشعارات أو تسجيل إزعاج |
| Create Issue | `CREATE_ISSUE` | إنشاء تذاكر ومهام جديدة داخل المشروع |
| Delete Issue | `DELETE_ISSUE` | حذف التذاكر نهائياً من المشروع |
| Link Issues | `LINK_ISSUE` | إنشاء وتعديل وحذف الروابط بين التذاكر |
| Override Visibility Restrictions | `READ_HIDDEN_STUFF` | تجاوز قيود الرؤية وقراءة المحتوى المقيد والمخفي |
| Read Issue | `READ_ISSUE` | قراءة الحقول العامة للتذكرة |
| Read Issue Private Fields | `PRIVATE_READ_ISSUE` | قراءة الحقول الخاصة والمقيدة داخل التذكرة |
| Update Issue | `UPDATE_ISSUE` | تعديل الحقول والبيانات العامة للتذكرة |
| Update Issue Private Fields | `PRIVATE_UPDATE_ISSUE` | تعديل الحقول الخاصة والمحمية داخل التذكرة |
| Update Watchers | `UPDATE_WATCHERS` | إضافة أو إزالة المتابعين (Watchers) من التذكرة |
| View Voters | `VIEW_VOTERS` | استعراض قائمة المصوتين لصالح التذكرة |
| View Watchers | `VIEW_WATCHERS` | استعراض قائمة المستخدمين المتابعين للتذكرة |

### 4. مرفقات التذاكر (Issue Attachment)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Add Attachment | `CREATE_ATTACHMENT_ISSUE` | إرفاق ملفات ومستندات بالتذكرة |
| Delete Attachment | `DELETE_ATTACHMENT_ISSUE` | حذف الملفات المرفقة بالتذكرة |
| Update Attachment | `UPDATE_ATTACHMENT_ISSUE` | تعديل بيانات المرفقات وخصائصها |

### 5. تعليقات التذاكر (Issue Comment)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Issue Comment | `CREATE_COMMENT` | إضافة تعليق جديد على التذكرة |
| Delete Issue Comment | `DELETE_COMMENT` | حذف التعليقات الذاتية للمستخدم |
| Delete Not Own and Permanent Comment Delete | `DELETE_NOT_OWN_COMMENT` | حذف تعليقات المستخدمين الآخرين أو الحذف النهائي |
| Read Issue Comment | `READ_COMMENT` | قراءة التعليقات الموجودة على التذكرة |
| Update Issue Comment | `UPDATE_COMMENT` | تعديل التعليقات الذاتية |
| Update Not Own Issue Comment | `UPDATE_NOT_OWN_COMMENT` | تعديل تعليقات المستخدمين الآخرين |

### 6. عناصر العمل والوقت (Issue Work Item / Time Tracking)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Not Own Work Item | `CREATE_NOT_OWN_WORK_ITEM` | إضافة سجلات وقت وعمل بالنيابة عن مستخدمين آخرين |
| Create Work Item | `CREATE_WORK_ITEM` | إضافة سجلات عمل ووقت ذاتية |
| Read Work Item | `READ_WORK_ITEM` | قراءة سجلات الوقت وتتبع العمل |
| Update Not Own Work Item | `UPDATE_NOT_OWN_WORK_ITEM` | تعديل سجلات عمل تعود لمستخدمين آخرين |
| Update Work Item | `UPDATE_WORK_ITEM` | تعديل سجلات العمل والوقت الذاتية |

### 7. المؤسسات (Organization)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Organization | `CREATE_ORGANIZATION` | إنشاء مؤسسات جديدة في النظام |
| Delete Organization | `DELETE_ORGANIZATION` | حذف مؤسسة من النظام |
| Read Organization | `READ_ORGANIZATION` | قراءة بيانات المؤسسة وإعداداتها |
| Update Organization | `UPDATE_ORGANIZATION` | تعديل إعدادات المؤسسة وتعيين المشاريع |

### 8. المشاريع (Project)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Project | `CREATE_PROJECT` | إنشاء مشروع جديد في النظام |
| Delete Project | `DELETE_PROJECT` | حذف مشروع كامل من النظام |
| Read Project Full | `READ_PROJECT` | قراءة بيانات المشروع التفصيلية والإدارية |
| Read Project Basic | `READ_PROJECT_BASIC` | قراءة البيانات الأساسية للمشروع (الاسم، المعرف) |
| Update Project | `UPDATE_PROJECT` | تعديل إعدادات المشروع وحقوله المخصصة |

### 9. النظام والتطبيقات منخفضة المستوى (System)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Low-level Admin Read | `ADMIN_READ_APP` | قراءة إعدادات التطبيقات والنظام منخفضة المستوى |
| Low-level Admin Write | `ADMIN_UPDATE_APP` | تعديل وتثبيت إعدادات التطبيقات والنظام على مستوى الإدارة |

### 10. المستخدمون والملفات الشخصية (User)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create User | `CREATE_USER` | إنشاء حسابات مستخدمين جديدة |
| Delete User | `DELETE_USER` | حذف أو حظر حسابات المستخدمين |
| Read User Full | `READ_USER` | قراءة الملف الشخصي الكامل للمستخدم (بما في ذلك البيانات الحساسة) |
| Read User Basic | `READ_USER_BASIC` | قراءة البيانات العامة والأساسية للمستخدم (الاسم، الصورة) |
| Update Self | `UPDATE_PROFILE` | تعديل المستخدم لملفه الشخصي وإعداداته الذاتية |
| Update User | `UPDATE_USER` | تعديل حسابات وبيانات المستخدمين الآخرين |

### 11. مجلدات المراقبة والوسوم وعمليات البحث (Watch Folder / Tags / Saved Searches)

| اسم الصلاحية في الواجهة | مفتاح الصلاحية البرمجي (Key) | الغرض والوصف |
| :--- | :--- | :--- |
| Create Tag or Saved Search | `CREATE_WATCH_FOLDER` | إنشاء وسوم (Tags) أو استعلامات بحث محفوظة |
| Delete Tag or Saved Search | `DELETE_WATCH_FOLDER` | حذف الوسوم وعمليات البحث المحفوظة |
| Edit Tag or Saved Search | `UPDATE_WATCH_FOLDER` | تعديل إعدادات الوسوم والبحوث المحفوظة |
| Share Custom View | `SHARE_WATCH_FOLDER` | مشاركة الوسوم والبحوث المحفوظة واللوحات مع مستخدمين آخرين |

---

## 5. ملاحظات تقنية واستنتاجات معمارية للمطورين

1. **التمييز بين `READ_USER` و `READ_USER_BASIC`:**
   - إذا كانت الودجة تحتاج فقط لعرض أسماء المستخدمين أو صورهم المصغرة (Avatar)، يُنصح بشدة بطلب `READ_USER_BASIC` بدلاً من `READ_USER`. طلب `READ_USER` سيحجب الودجة عن المستخدمين العاديين وضيوف المشاريع الذين لا يملكون صلاحيات إدارية على ملفات المستخدمين.
2. **التمييز بين `READ_PROJECT` و `READ_PROJECT_BASIC`:**
   - لعرض قائمة المشاريع فقط في قائمة منسدلة داخل الودجة، يكفي التحقق من `READ_PROJECT_BASIC`. بينما `READ_PROJECT` يتطلب صلاحيات أعلى للوصول لإعدادات المشروع وحقوله.
3. **التكامل مع نموذج الأدوار (RBAC):**
   - مطور التطبيق لا يتعامل مع "الأدوار" (Roles) في ملف البيان، بل يتعامل حصراً مع "مفاتيح الصلاحيات" (Permission Keys). هذا الفصل يضمن مرونة التطبيق؛ إذ يمكن للمؤسسة تخصيص الأدوار كما تشاء دون كسر قواعد ظهور الودجات.
4. **تجنب التعارض مع الصلاحيات المتأصلة (Inherent Rights):**
   - يجب الانتباه إلى أن مستخدم التذكرة الذي أنشأها (Reporter) يمتلك حقوقاً متأصلة لتعديلها. فإذا صُممت ودجة تتطلب حصراً `UPDATE_ISSUE`، فقد تُحجب عن كاتب التذكرة رغم قدرته الفعلية على تعديل تذكرته، لذا يجب ضبط مصفوفة `permissions` بعناية لتشمل البدائل المناسبة مثل `["UPDATE_ISSUE", "CREATE_ISSUE"]`.

---

## 6. الارتباط مع أدلة المشروع التوثيقية الأخرى

- [دليل مرجع صلاحيات يوتراك الأساسي](file:///home/osm/StudioProjects/permissions_youtrack/docs/permissions/youtrack_permissions_reference.md)
- [دليل مستويات النطاقات (Scope Levels)](file:///home/osm/StudioProjects/permissions_youtrack/docs/permissions/scope_levels_guide.md)
- [دليل الصلاحيات المتأصلة (Inherent Permissions)](file:///home/osm/StudioProjects/permissions_youtrack/docs/permissions/inherent_permissions_guide.md)
- [دليل وحدات يوتراك البرمجية (YouTrack Modules)](file:///home/osm/StudioProjects/permissions_youtrack/docs/permissions/youtrack_modules_guide.md)
- [دليل أنواع الكيانات (Entities Guide)](file:///home/osm/StudioProjects/permissions_youtrack/docs/permissions/youtrack_entities_guide.md)
