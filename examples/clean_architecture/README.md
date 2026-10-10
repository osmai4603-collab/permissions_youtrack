# Clean Architecture & RBAC Engine Integration Showcase

يوضح هذا المشروع المصغر (`examples/clean_architecture`) كيفية تطبيق **معمارية المعمارية النظيفة (Clean Architecture)** للمهندس روبرت مارتن (Uncle Bob) وكيفية دمج **محرك الأذونات والصلاحيات المتقدم (RBAC/ABAC Engine)** داخله بأعلى معايير فصل المسؤوليات (Separation of Concerns) ومبدأ عكس التبعية (Dependency Inversion Principle - DIP).

---

## 🏗️ هيكلية الطبقات الأربع (The 4 Clean Architecture Layers)

تم تنظيم المشروع داخل 4 مجلدات/طبقات رئيسية، تعكس قواعد تدفق الاعتمادية (Dependency Rule) من الخارج نحو الداخل:

```text
examples/clean_architecture/
├── 1_domain/                 # [الطبقة 1: النواة والمنافذ] كيانات النطاق والعقود النقية
│   ├── issue.go              # كيان المهمة Issue Entity (لا يعتمد على أي مكتبة خارجية)
│   └── ports.go              # المنافذ (IssueRepository, Authorizer Interfaces)
│
├── 2_usecases/               # [الطبقة 2: حالات الاستخدام] منطق تطبيق الأعمال
│   ├── create_issue.go       # حالة استخدام إنشاء مهمة (CreateIssueUseCase)
│   └── delete_issue.go       # حالة استخدام حذف مهمة مع فحص الملكية الملازمة (DeleteIssueUseCase)
│
├── 3_adapters/               # [الطبقة 3: المحولات وبوابات الواجهة]
│   ├── mem_repository.go     # محول التخزين في الذاكرة (Outbound Adapter لمستودع المهام)
│   ├── http_middleware.go    # محول الحراسة والوسيط الأمني (Edge PEP Middleware)
│   └── issue_controller.go   # محول تسليم HTTP وتحويل الـ DTOs إلى أوامر
│
├── app_test.go               # [الطبقة 4: الأطر والربط واختبارات التكامل الشاملة]
└── README.md                 # هذا الدليل التوضيحي
```

---

## 🎯 أين وكيف يتكامل محرك الـ RBAC مع الطبقات؟

| الطبقة في المعمارية النظيفة | مكون RBAC المناظر | الموقع في الكود | الوظيفة والمسؤولية |
| :--- | :--- | :--- | :--- |
| **1. Domain Core** | **PDP Engine & Graphs** | `internal/services/rbac` | النماذج الأساسية (`Role`, `Permission`, `AccessRequest`)، خوارزميات الـ DAG، وفحص قيود الفصل بين الواجبات (SoD). |
| **1. Domain Ports** | **Authorizer Interface** | `1_domain/ports.go` | تعريف منفذ التخويل `Authorizer` لتمكين حقن التبعية ومنع ارتباط حالات الاستخدام بأي محرك أمني محدد. |
| **2. Use Cases** | **Internal PEP** | `2_usecases/delete_issue.go` | **الدفاع في العمق (Defense-in-Depth):** تنفيذ فحص الصلاحية الدقيق وحقن السياق والملكية الفطرية (`author_id == subject_id`). |
| **3. Interface Adapters** | **Edge PEP Middleware** | `3_adapters/http_middleware.go` | **الفحص السريع عند الحافة (Fast-Fail):** استخراج الهوية والرمز، التحقق من الجلسة والصلاحيات العامة، ورفض الطلب (401/403) قبل إجهاد طبقات التطبيق. |
| **4. Frameworks & Composition** | **Dependency Injection Wire-up** | `app_test.go` | تركيب المحرك وربط المستودعات وضبط قواعد السياق وسيناريوهات تصعيد الامتيازات اللحظية (JIT Elevation). |

---

## 💡 سيناريوهات التكامل المطبقة في الاختبارات (`app_test.go`)

### 1. الوراثة الهرمية (Hierarchical Inheritance)

- دور `DEVELOPER` يمتلك صلاحيات `CREATE_ISSUE`, `READ_ISSUE`, `DELETE_ISSUE` (المقيدة بالسياق).
- دور `PROJECT_ADMIN` يرث دور `DEVELOPER` ويمتلك صلاحية الحذف المطلق في نطاق المشروع.

### 2. الصلاحية الفطرية الملازمة (Inherent Ownership)

- المستخدم **Bob** يملك دور مطور فقط ولا يملك صلاحية مدير مشروع.
- عندما حاول Bob حذف مهمة كتبتها **Alice**، رُفض طلبه برمز `403 Forbidden`.
- عندما أنشأ Bob مهمته الخاصة، سمح له النظام بحذفها فوراً برمز `204 No Content` لأن محرك الـ PDP قام بمطابقة `InherentOwnershipEvaluator` بنجاح (المؤلف يملك صلاحية إدارة ما يكتبه تلقائياً).

### 3. تصعيد الامتيازات المؤقت اللحظي (Just-In-Time Elevation)

- المستخدم **Dave** يملك جلسة عادية كمطور ولا يمكنه حذف مهمة نظام حساسة (`403 Forbidden`).
- يتم استدعاء ترقية امتيازه مؤقتاً لدور `PROJECT_ADMIN` لمدة 40 ميلي ثانية عبر `ActivateRoleInSession`.
- خلال النافذة الزمنية، يتم قبول طلب الحذف بنجاح (`204 No Content`).
- بمجرد انقضاء المهلة الزمنية، يقوم محرك الجلسات بالإخلاء التلقائي (Auto-eviction)، ويعود الطلب التالي ليرفض بـ `403 Forbidden` تلقائياً دون الحاجة لتدخل يدوي.

---

## 🧪 تشغيل الاختبارات الشاملة (Running the Tests)

لتشغيل اختبارات التكامل وكشف أي سباق بيانات:

```bash
go test -v -race ./examples/clean_architecture/...
```
