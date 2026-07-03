# FastSub-Scanner ⚡

أداة سطر أوامر (CLI) عالية السرعة مبنية بلغة **Go**، مخصصة للمطورين ومختبري الاختراق للتحقق بسرعة من حالة النطاقات الفرعية (Subdomains) — بدون تعقيد الأدوات الضخمة.

> A blazing-fast Go CLI tool to check subdomain liveness, HTTP status codes, and server headers using massive concurrency.

---

## ✨ المميزات (Features)

- **⚡ فحص متزامن حقيقي (True Concurrency)** — باستخدام Goroutines + Worker Pool، يمكن فحص آلاف النطاقات في ثوانٍ.
- **🔍 كشف تلقائي للبروتوكول** — يجرب HTTPS أولاً ثم HTTP تلقائيًا.
- **📊 معلومات أمنية سريعة** — HTTP Status Code + Server Header لكل نطاق فعّال.
- **📁 استيراد من ملف نصي** — اقرأ قائمة نطاقات غير محدودة من ملف `.txt`.
- **💾 تصدير النتائج** — إلى `JSON` أو `CSV` بضغطة واحدة، جاهزة للدمج مع أدوات أخرى.
- **🎛️ تحكم كامل** — عدد الخيوط (Concurrency)، المهلة الزمنية (Timeout)، وفلترة العرض.

---

## 📦 التثبيت (Installation)

```bash
git clone https://github.com/yourusername/fastsub-scanner.git
cd fastsub-scanner
go build -o fastsub-scanner main.go
```

يتطلب [Go 1.20+](https://go.dev/dl/) مثبتًا على جهازك.

---

## 🚀 طريقة الاستخدام (Usage)

### 1. جهّز ملف النطاقات

أنشئ ملف `domains.txt` وضع فيه نطاقًا واحدًا في كل سطر:

```
github.com
pypi.org
npmjs.com
target-subdomain.example.com
```

### 2. شغّل الأداة

```bash
./fastsub-scanner -l domains.txt
```

### الخيارات المتاحة (Flags)

| الخيار         | الوصف                                              | الافتراضي |
|----------------|-----------------------------------------------------|-----------|
| `-l`           | مسار ملف النطاقات (إلزامي)                          | —         |
| `-c`           | عدد الفحوصات المتزامنة (Concurrency)                | `50`      |
| `-t`           | المهلة الزمنية بالثواني لكل طلب                      | `5`       |
| `-json`        | مسار ملف لتصدير النتائج بصيغة JSON                   | —         |
| `-csv`         | مسار ملف لتصدير النتائج بصيغة CSV                    | —         |
| `-alive-only`  | عرض النطاقات الفعّالة فقط في الطرفية                 | `false`   |

### مثال متقدم

```bash
./fastsub-scanner -l domains.txt -c 100 -t 3 -json results.json -csv results.csv -alive-only
```

---

## 📸 مثال على الناتج (Sample Output)

تشغيل حقيقي على 5 نطاقات (تضمّن نطاقًا غير موجود لاختبار حالة "Down"):

```
$ ./fastsub-scanner -l test_domains.txt -c 10 -t 5 -json results.json -csv results.csv

[*] تم تحميل 5 نطاق للفحص
[*] عدد الخيوط المتزامنة: 10 | المهلة: 5s

[Alive] pypi.org                            | 200 | https  | Server: gunicorn
[Alive] github.com                          | 200 | https  | Server: github.com
[Down]  this-domain-should-not-exist-xyz123.com | Get "http://...": no such host
[Alive] raw.githubusercontent.com           | 200 | https  | Server: github.com
[Alive] npmjs.com                           | 403 | https  | Server: cloudflare

[*] اكتمل الفحص في 500ms
[*] النتيجة: 4 فعّال / 5 إجمالي
[+] تم حفظ النتائج في: results.json
[+] تم حفظ النتائج في: results.csv
```

⏱️ **5 نطاقات تم فحصها بالكامل خلال 500 ميلي ثانية فقط** بفضل الفحص المتزامن.

### مثال ناتج JSON

```json
[
  {
    "domain": "github.com",
    "alive": true,
    "status_code": 200,
    "scheme": "https",
    "server": "github.com"
  },
  {
    "domain": "this-domain-should-not-exist-xyz123.com",
    "alive": false,
    "error": "dial tcp: lookup ... no such host"
  }
]
```

### مثال ناتج CSV

| domain | alive | status_code | scheme | server | error |
|--------|-------|-------------|--------|--------|-------|
| github.com | true | 200 | https | github.com | |
| pypi.org | true | 200 | https | gunicorn | |

---

## 🧠 كيف تعمل الأداة تقنيًا (How it works)

تستخدم الأداة نمط **Worker Pool** بدلاً من إطلاق Goroutine لكل نطاق بلا حدود — هذا يمنع استنزاف الموارد عند فحص قوائم ضخمة (10,000+ نطاق) ويجعل التحكم بعدد الاتصالات المتزامنة (`-c`) دقيقًا وقابلاً للتوقع، عكس الكود الأولي البسيط الذي كان يطلق كل الفحوصات دفعة واحدة بدون تحكم.

---

## ⚠️ إخلاء مسؤولية (Disclaimer)

هذه الأداة مخصصة **للاستخدام الأخلاقي فقط** — اختبار اختراق مصرّح به، بحث أمني على أصولك الخاصة، أو أهداف تعليمية. لا تستخدمها على أنظمة لا تملك إذنًا صريحًا بفحصها. المستخدم يتحمل المسؤولية الكاملة عن أي استخدام غير قانوني.

---

## 📄 الترخيص (License)

MIT License — استخدم، عدّل، وزّع بحرية.
