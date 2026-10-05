# حقل نوع المجموعة (Group Type) في YouTrack — User Group مقابل Customer Group

**نوع المستند:** بحث مرجعي من المصادر الرسمية
**المصدر الرئيسي:** JetBrains YouTrack Server Documentation 2026.2 + YouTrack Developer Portal
**تاريخ البحث:** 2026-10-05
**الموضوع:** الحقل الذي يحدد نوع المجموعة في كيان `UserGroup` — وهل هو متاح في REST API أم في Workflow API فقط

---

## الخلاصة التنفيذية

| السؤال | الإجابة |
| --- | --- |
| ما اسم الحقل؟ | `isCustomerGroup` |
| ما نوعه؟ | `Boolean` |
| في أي واجهة برمجية؟ | **Workflow/Scripting API فقط** |
| متاح منذ أي إصدار؟ | **2026.2** |
| قابل للكتابة؟ | **لا — للقراءة فقط** |
| هل يوجد مقابل له في REST API؟ | **لا — غير معلن في `/api/groups`** |

> **النتيجة الحاسمة:** YouTrack لا يوفّر حقل نوع المجموعة في واجهة REST API. الحقل الرسمي الوحيد هو `UserGroup.isCustomerGroup` في Workflow API، وهو علَم ثنائي (boolean) لا حقل نصي مقيّد (enum)، وهو **للقراءة فقط**.

---

## 1. الحقل الرسمي في Workflow API

### 1.1 `UserGroup.isCustomerGroup`

النص الرسمي من [UserGroup | Developer Portal](https://www.jetbrains.com/help/youtrack/devportal/v1-UserGroup.html):

> *"isCustomerGroup | Boolean | Read-only. If this group is a helpdesk customer group, this property is `true`. Available since 2026.2"*

### 1.2 الحقل المرافق: `customerGroupProjects`

> *"customerGroupProjects | Set of Project | Read-only. The set of helpdesk projects where this group is configured as a customer group. For groups that are not customer groups, this property is an empty set. Available since 2026.2"*

هذا الحقل يكمّل الأول: `isCustomerGroup` يجيب على السؤال **هل هي مجموعة عملاء؟**، و `customerGroupProjects` يجيب على السؤال **في أي مشاريع Helpdesk؟**

### 1.3 تأكيد الإصدار من سجل التغييرات

من [Change Log | Developer Portal](https://www.jetbrains.com/help/youtrack/devportal/Workflow-Change-Log.html)، قسم **2026.2**:

> - *"Added property `UserGroup.customerGroupProjects`"*
> - *"Added property `UserGroup.isCustomerGroup`"*

لاحظ أن الإصدار **2026.2** أضاف أيضاً `User.type` و `UserType` و `Issue.customerGroups`. هذا يعني أن **نوع المجموعة ككيان أول-class** ميزة حديثة نسبياً (2026.2)، وليست موجودة في الإصدارات الأسبق.

### 1.4 الوراثة — تنطبق أيضاً على `ProjectTeam`

الحقل نفسه موروث في كيان `ProjectTeam`، أي أن فرق المشاريع تحمل الخاصية (قيمتها `false` دائماً عملياً):

> *"isCustomerGroup | Boolean | Read-only. If this group is a helpdesk customer group, this property is `true`. Available since 2026.2"*
> — [ProjectTeam | Developer Portal](https://www.jetbrains.com/help/youtrack/devportal/v1-ProjectTeam.html)

---

## 2. الفرق بين الواجهتين البرمجيتين — الفجوة الحرجة

هذه هي النقطة الأهم عملياً في هذا البحث، لأنها تؤثر على بنية المشروع الحالية.

### 2.1 ما تعلنه REST API لـ `/api/groups`

جدول السمات الكامل المعلن في [User Groups | Developer Portal](https://www.jetbrains.com/help/youtrack/devportal/resource-api-groups.html):

| Field | Type | الوصف | للقراءة فقط؟ |
| --- | --- | --- | --- |
| `id` | String | معرّف المجموعة | نعم |
| `name` | String | اسم المجموعة | لا |
| `ringId` | String | المعرّف في Hub | نعم، قد يكون `null` |
| `usersCount` | Long | عدد المستخدمين | نعم |
| `icon` | String | رابط شعار المجموعة | نعم، قد يكون `null` |
| `allUsersGroup` | Boolean | هل هي مجموعة All Users | نعم |
| `users` | Array of Users | جميع المستخدمين شاملاً العابرين | نعم |

**لا يوجد أي حقل نوع في هذا الجدول.** لا `type`، ولا `isCustomerGroup`، ولا `customerGroupProjects`.

> **الخلاصة:** لا يمكن تمييز مجموعة العملاء عن مجموعة المستخدمين عبر REST API `/api/groups`.

### 2.2 لماذا هذا مهم لمشروعنا

الجدول أعلاه يطابق **حرفياً** أعمدة جدول `groups` في `migrations/000003_groups.sql` في هذا المشروع:

```sql
id, name, ring_id, users_count, icon, all_users_group, auto_join
```

باستثناء `auto_join` و `created_at` / `updated_at` (وهي أعمدة محلية). غياب عمود النوع في المخطط الحالي **متسق تماماً مع ما تعلنه REST API** — أي أن نقصه ليس نقصاً في بحثنا بل انعكاس لحدود الواجهة الرسمية.

### 2.3 `$type` لا يحمل الإجابة أيضاً

في استجابات REST، الحقل `$type` يُميّز بين فئات الكيانات:

```text
"NestedGroup", "AllUsersGroup", "RegisteredUsersGroup", "ProjectTeam"
```

من [api-entity-UserGroup](https://www.jetbrains.com/help/youtrack/devportal/api-entity-UserGroup.html)، التسلسل الهرمي الكامل لكيان `UserGroup`:

```text
UserGroup
├── AllUsersGroup
├── RegisteredUsersGroup
├── ProjectTeam
└── NestedGroup
```

**لا توجد فئة `CustomerGroup` في هذا التسلسل.** مجموعة العملاء هي `NestedGroup` عادي، يتميّز بخاصية وليس بنوع كيان.

---

## 3. مصدر التمييز البديل: واجهة البحث (Search UI)

في غياب حقل API، توفّر YouTrack وسيلة تمييز رسمية عبر واجهة البحث في صفحة المجموعات.

من [Search for Groups | YouTrack Server Documentation](https://www.jetbrains.com/help/youtrack/server/search-groups.html):

> *"Is customer group | Returns groups that are or are not customer groups for helpdesk projects."*

### جدول عوامل التصفية المعتمدة بالكامل

| العامل | الوصف |
| --- | --- |
| `Auto-join` | المجموعات التي خيار الانضمام التلقائي مفعّل فيها |
| `Has subgroups` | المجموعات التي تحتوي مجموعة فرعية واحدة على الأقل |
| `Is project team` | هل المجموعة فريق مشروع أم لا |
| **`Is customer group`** | **هل المجموعة مجموعة عملاء لمشروع Helpdesk أم لا** |
| `Member` | المجموعات التي المستخدم المحدد عضو فيها |
| `Parent group` | المجموعات الفرعية للمجموعة الأب المحددة |
| `Permission` | المجموعات التي مُسندة إليها صلاحية محددة |
| `Project team` | المجموعات المضافة إلى فريق المشروع المحدد |
| `Requires 2FA` | المجموعات التي إلزام 2FA مفعّل فيها |
| `Role` | المجموعات التي مُسندة إليها دور محدد |

**محصلة التغطية:**

| مستوى التمييز | متاح؟ |
| --- | --- |
| Workflow API | نعم — `isCustomerGroup` |
| REST API `/api/groups` | **لا** |
| Search UI | نعم — `Is customer group` |
| `$type` discriminator | **لا** |
| Schema قاعدة البيانات في مشروعنا | **لا** |

---

## 4. قواعد النوع — لماذا لا يوجد حقل enum

قرار YouTrack باستخدام **boolean** بدلاً من حقل نصي (`USER_GROUP` / `CUSTOMER_GROUP`) ليس تبسيطاً عشوائياً؛ سببه أن عدد الأنواع **ثابت ومقصود**:

> *"The group type is set when the group is created. Existing user groups can't be converted into customer groups, and customer groups can't be converted into user groups."*
>
> — [Helpdesk Customer Groups | YouTrack Server Documentation](https://www.jetbrains.com/help/youtrack/server/helpdesk-customer-groups.html)

بما أن مجموع الأنواع محصور في قيمتين لا ثالثة لهما، فإن `isCustomerGroup = false` تعني ضمناً "User group". **لا يوجد حالة ثالثة** تحتاج تمثيلاً.

### القيود الرسمية لمجموعات العملاء

> *"Customer groups don't grant roles or permissions. They only provide helpdesk ticket access when the group is used in helpdesk-specific settings."*
>
> - *"Roles can't be assigned to customer groups."*
> - *"The group type is set when the group is created. Existing user groups can't be converted into customer groups, and customer groups can't be converted into user groups."*
> - *"Groups synchronized from an external identity provider aren't supported as customer groups."*
> - *"Customer groups can't be nested under other groups and can't contain subgroups."*
> - *"Customer groups can't be added to project teams or used as values in group custom fields."*
> - *"Customer groups can't be used in visibility settings for articles, comments, attachments, agile boards, reports, dashboards, or standard issues in non-helpdesk projects."*

### إنشاء المجموعة — نقطة تحديد النوع

> *"4. Select the group type.*
>
> - *User group creates a standard group that can be assigned roles and used to manage access for its members.*
> - *Customer group creates a helpdesk group whose members can view and comment on tickets shared with the group. Customer groups can't be assigned roles or nested under other groups."*
>
> — [Create a Group | YouTrack Server Documentation](https://www.jetbrains.com/help/youtrack/server/create-user-group.html)

النوع يُحدَّد **لحظة الإنشاء فقط**، ومسار الإنشاء في واجهة Helpdesk هو:

`Project Settings > Customer Groups > Add customer group > New customer group`

### ما الذي يبقى متاحاً لمجموعة العملاء؟

> *"Customer groups have the same global settings for members, auto-join, two-factor authentication, and visibility of the group. Role assignment, nesting, project team membership, and audit links are not available for customer groups."*
>
> — [Edit Basic Group Settings | YouTrack Server Documentation](https://www.jetbrains.com/help/youtrack/server/edit-basic-settings-of-a-group.html)

| الإعداد | متاح لمجموعة العملاء؟ |
| --- | --- |
| Members | نعم |
| Auto-join | نعم (عبر مطابقة نطاق البريد) |
| Two-factor authentication | نعم |
| Visible to / Updatable by | نعم |
| Role assignment | **لا** |
| Nesting | **لا** |
| Project team membership | **لا** |
| Audit link | **لا** |
| Auto-share (خاص بالمشروع) | نعم |

### الشارة البصرية

> *"Groups marked with the Helpdesk badge are customer groups."*
> — [Groups | YouTrack Server Documentation](https://www.jetbrains.com/help/youtrack/server/manage-user-groups.html)

الشارة البصرية هي المؤشر الوحيد المتاح للمستخدم في الواجهة الرسومية، وهي **غير قابلة للاستخراج برمجياً**.

---

## 5. مصفوفة مقارنة: المجموعات العادية مقابل مجموعات العملاء

| المعيار | User Group | Customer Group |
| --- | --- | --- |
| الغرض | إدارة الوصول والأدوار عبر المشاريع | مشاركة تذاكر الـ Helpdesk مع جهة خارجية |
| إسناد الأدوار | مدعوم | **محظور** |
| التداخل (nesting) | مدعوم | **محظور** |
| كأبناء أو أولياء | مدعوم | **محظور** |
| فريق المشروع | مدعوم | **محظور** |
| حقول المجموعات المخصصة | مدعوم | **محظور** |
| إعدادات الرؤية العامة | مدعوم | **محظور** (خارج مشاريع Helpdesk) |
| مزامنة مزود الهوية | مدعومة | **غير مدعومة** |
| تدقيق الأحداث (Audit) | مدعوم | **غير متاح** |
| أنواع الحسابات المسموح عضويتها | أي حساب | مُبلّغون، مستخدمون عاديون، وكلاء |
| `isCustomerGroup` | `false` | `true` |
| `customerGroupProjects` | مجموعة فارغة | مجموعة مشاريع Helpdesk |

### ملاحظة أمنية: مصفوفة الاستثناءات

مجموعة العملاء **استثناء جوهري** عن قاعدة "الصلاحيات تأتي عبر الأدوار". هي قناة صلاحيات ثانوية لا تدخل في دورة الأدوار وتفتح فقط:

- عرض التذاكر المشتركة مع المجموعة
- إضافة تعليقات عامة على تلك التذاكر
- اشتراك الإشعارات "Shared with my customer groups"

> *"Customer groups don't grant roles or permissions. They only provide helpdesk ticket access when the group is used in helpdesk-specific settings."*

أي نموذج تصميم يفترض أن **كل** مجموعة في `groups`_visible إلى مسار الدور يحتاج استثناءً صريحاً لمجموعات العملاء.

---

## 6. التوافق مع جدول `groups` الحالي

### 6.1 ما كان ناقصاً قبل التحديث

الجدول الأصلي كان يطابق جدول سمات REST API المعلن حرفياً، وبالتالي كان يخلو من كل ما لا تعلنه REST:

| الفجوة | مرجع رسمي |
| --- | --- |
| نوع المجموعة | `UserGroup.isCustomerGroup` — Workflow API فقط |
| مشاريع Helpdesk المرتبطة | `UserGroup.customerGroupProjects` |
| الوصف | `Edit Basic Group Settings` → `Description` |
| شجرة التداخل | `NestedGroup.parentGroup` / `subGroups` |
| العضويات | `NestedGroup.ownUsers` / `UserGroup.users` |
| Visible to / Updatable by | `NestedGroup.viewers` / `updaters` |
| إلزام 2FA | `NestedGroup.requireTwoFactorAuthentication` |
| نطاق الانضمام التلقائي | `NestedGroup.autoJoinDomain` |
| علامة Registered Users | كيان `RegisteredUsersGroup` |

### 6.2 الحالة بعد التحديث

أضيفت الفجوات أعلاه إلى `migrations/000003_groups.sql` على النحو التالي:

| Official | العنصر في المشروع | ملاحظات |
| --- | --- | --- |
| `isCustomerGroup` | `groups.is_customer_group` | Boolean، افتراضي FALSE |
| `customerGroupProjects` | `group_customer_projects` | علاقة متعدد-أطراف، مع `auto_share` الخاص بكل مشروع |
| `Description` | `groups.description` | — |
| `parentGroup` | `groups.parent_group_id` | مرجع ذاتي، `RESTRICT` عند الحذف عمداً |
| `ownUsers` | `group_members` | العضويات المباشرة فقط |
| `users` (transitive) | `view group_members_transitive` | Union على الشجرة مع حماية من الدورات |
| `viewers` / `updaters` | `group_visibility` | هدف متعدد الأشكال (مستخدم أو مجموعة) |
| `requireTwoFactorAuthentication` | `groups.require_two_factor_authentication` | — |
| `autoJoinDomain` | `groups.auto_join_domain` | آلية الاعتماد الأساسية لمجموعات العملاء |
| `RegisteredUsersGroup` | `groups.registered_users_group` | حصر متبادل مع `all_users_group` |

### 6.3 اتجاه الوراثة في `group_members_transitive`

العضوية تنتشر **نازلاً في الشجرة** لا صاعداً. السبب أن التوثيق ينص على أن الابن يرث أدوار الأب:

> *"The group also inherits roles that are assigned to a parent group. For example, roles that are assigned to the All Users group are available to all the groups that are nested under this group."*

هذه الوراثة تكون صحيحة فقط ما دام كل عضو في الابن عضواً في الأب أيضاً، أي أن الابن **مجموعة جزئية** من الأب. لذلك، مستخدم انتقل إلى مجموعة أعلاه في الشجرة ليس بالضرورة عضواً في المجموعة الأعلى. المستخدمون العابرون لمجموعة = أعضاؤها المباشرون + أعضاء كل المجموعات المتداخلة تحتها.

**الاستثناء:** كيان `ProjectTeam` ما زال غير ممثَّل — قرار متعمد لتفادي عمود بلا مستهلك، لأنه كيان مستقل في YouTrack ويحمل مرجع مشروع إلزامياً.

### 6.4 مصادر لا تزال غير قابلة للمزامنة

| العنصر | السبب |
| --- | --- |
| `is_customer_group` | غير معلن في `/api/groups` |
| `group_customer_projects` | خاصية Workflow API منذ 2026.2 |

كلاهما يحتاج Workflow API أو تكامل apps للملء. أما بقية الأعمدة فمسحوبة من REST مباشرة.

**التوصية:** أبقِ `groups.users_count` كقيمة مخزَّنة، وأعد اشتقاقه من `group_members_transitive` عند الحاجة بدل الوثوق بمزامنة عدّاد YouTrack.

---

## 7. الفارق بين نوع المجموعة ونوع المستخدم

YouTrack أضافت في الإصدار نفسه (2026.2) تصنيفاً موازياً **للمستخدم**، وهو تمييز مفيد لتجنب الالتباس:

| الكيان | الحقل | القيم |
| --- | --- | --- |
| `User` | `type` | نوع المستخدم (كيان `UserType`) |
| `UserType` | ثوابت | `AGENT`, `REPORTER`, `STANDARD_USER` |
| `UserType` | `typeName` | اسم النوع كنص |
| `UserGroup` | `isCustomerGroup` | `true` / `false` |

**التصنيفان مستقلان تماماً:** مجموعة العملاء يمكن أن تضم **مُبلّغين ومستخدمين عاديين ووكلاء** في آن واحد:

> *"Customer groups are based on the standard group model, but they are scoped to helpdesk use cases. They are marked with a Helpdesk badge in the global groups list and can include reporters, standard users, and agents."*

لاحظ التناقض الظاهري في `UserType`: التوثيق يعرّف `STANDARD_USER` في Workflow API بينما تبويب REST يسميه `userType`. هذه طبقة تسوية، وليست تعارضاً في النموذج.

---

## 8. أسئلة مفتوحة تحتاج تحققاً ميدانياً

| السؤال | الحالة |
| --- | --- |
| هل يقبل `/api/groups?fields=isCustomerGroup` حقلاً غير معلن ويعيده؟ | **غير موثق** — لم يُعلن في جدول السمات |
| هل يعمل `?query=Is customer group: true` عبر REST؟ | **غير موثق** — العامل موثق لواجهة Search UI فقط |
| هل يمكن قراءة النوع من `/hub/api/rest/usergroups`؟ | **غير موثق** — لا يوجد حقل نوع معلن في واجهة Hub |
| هل `isCustomerGroup` متاح في `NestedGroup` أم `UserGroup`؟ | في `UserGroup` (الموروث) — `NestedGroup` لا يسرده في جدول سماته |
| هل يعمل البحث في إصدارات أقدم من 2026.2؟ | الخاصية غير موجودة قبل 2026.2 |

**التوصية:** التحقق من هذه النقاط يتطلب تقييماً حيّاً على خادم YouTrack 2026.2، وهو خارج نطاق البحث في المصادر الرسمية.

---

## 9. المصادر الرسمية

| # | المصدر | الرابط | الاستخدام |
| --- | --- | --- | --- |
| 1 | UserGroup \| Developer Portal | <https://www.jetbrains.com/help/youtrack/devportal/v1-UserGroup.html> | تعريف `isCustomerGroup` و `customerGroupProjects` |
| 2 | User Groups \| REST API | <https://www.jetbrains.com/help/youtrack/devportal/resource-api-groups.html> | جدول سمات `/api/groups` (بدون حقل نوع) |
| 3 | api-entity-UserGroup | <https://www.jetbrains.com/help/youtrack/devportal/api-entity-UserGroup.html> | التسلسل الهرمي للكيانات |
| 4 | ProjectTeam \| Developer Portal | <https://www.jetbrains.com/help/youtrack/devportal/v1-ProjectTeam.html> | وراثة الخاصية |
| 5 | Change Log | <https://www.jetbrains.com/help/youtrack/devportal/Workflow-Change-Log.html> | تأكيد إصدار 2026.2 |
| 6 | Helpdesk Customer Groups | <https://www.jetbrains.com/help/youtrack/server/helpdesk-customer-groups.html> | القيود الرسمية الكاملة |
| 7 | Create a Group | <https://www.jetbrains.com/help/youtrack/server/create-user-group.html> | نقطة تحديد النوع |
| 8 | Edit Basic Group Settings | <https://www.jetbrains.com/help/youtrack/server/edit-basic-settings-of-a-group.html> | ما هو متاح وما هو محظور |
| 9 | Search for Groups | <https://www.jetbrains.com/help/youtrack/server/search-groups.html> | عامل `Is customer group` |
| 10 | Groups | <https://www.jetbrains.com/help/youtrack/server/manage-user-groups.html> | الشارة البصرية Helpdesk |
| 11 | Helpdesk Features for Standard Users | <https://www.jetbrains.com/help/youtrack/server/helpdesk-standard-users.html> | التمييز بين المجموعات |

---

## خلاصة معمارية

1. **الحقل الرسمي هو `UserGroup.isCustomerGroup`** (Boolean، للقراءة فقط، منذ 2026.2) في Workflow API.
2. **لا يوجد حقل نوع في REST API** `/api/groups` — الفجوة بين الواجهتين حقيقية وموثقة.
3. **`$type` لا يكشف النوع** — لا فئة `CustomerGroup` في التسلسل الهرمي للكيانات.
4. **النوع ثابت عند الإنشاء** — لا تحويل بين النوعين،(boolean وليس enum لأن مجموع الأنواع مقصور على قيمتين).
5. **مجموعة العملاء استثناء معماري** عن قاعدة "الصلاحيات عبر الأدوار"، ولا يمكن أن تحمل أدواراً أو تُدرّج أو تُضاف لفرق مشاريع.
6. **بالنسبة للمشروع:** عمود `type` على `groups` لا يمكن مزامنته من REST API؛ يلزم مصدر بديل (Workflow API 2026.2+، أو جدول ارتباط منفصل).
