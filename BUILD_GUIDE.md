# دليل بناء وتوزيع FastSub-Scanner (Build & Release Guide)

هذا الدليل يشرح كيف تبني ملفات تنفيذية (Binaries) جاهزة للأنظمة الثلاثة، تصلح للبيع مباشرة للمستخدم النهائي.

---

## 1. المتطلبات (على جهاز البناء فقط)

تحتاج **أنت فقط كمطوّر** إلى تثبيت Go على جهازك لتبني الملفات:

- ثبّت Go 1.20 أو أحدث من [go.dev/dl](https://go.dev/dl/)
- تأكد من التثبيت:
  ```bash
  go version
  ```

⚠️ **مهم جدًا:** المستخدم النهائي (المشتري) **لا يحتاج تثبيت Go إطلاقًا** — فقط يشغّل الملف التنفيذي مباشرة. هذا هو معنى "Cross-Compilation": تبني أنت على جهازك ملفات تعمل على أي نظام، دون الحاجة لامتلاك ذلك النظام أو تثبيت Go عليه.

---

## 2. البناء اليدوي (لفهم الآلية)

لو أردت بناء نظام واحد يدويًا لفهم الفكرة، فالمتغيرات الأساسية هي `GOOS` (نظام التشغيل) و `GOARCH` (المعمارية):

```bash
# مثال: بناء نسخة Windows من على أي جهاز (حتى لو كنت على Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o fastsub-scanner-windows-amd64.exe main.go

# مثال: بناء نسخة macOS لأجهزة Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o fastsub-scanner-macos-arm64 main.go

# مثال: بناء نسخة Linux عادية
GOOS=linux GOARCH=amd64 go build -o fastsub-scanner-linux-amd64 main.go
```

الجدول المرجعي لأشهر التوليفات:

| النظام المستهدف          | GOOS      | GOARCH  |
|---------------------------|-----------|---------|
| Windows 64-bit             | `windows` | `amd64` |
| Linux 64-bit (Intel/AMD)   | `linux`   | `amd64` |
| Linux ARM (سيرفرات/Raspberry Pi) | `linux`   | `arm64` |
| macOS Intel                | `darwin`  | `amd64` |
| macOS Apple Silicon         | `darwin`  | `arm64` |

---

## 3. الأتمتة (الطريقة الموصى بها) ⚡

بدلاً من تكرار الأمر 5 مرات، استخدم أحد الملفات الجاهزة المرفقة:

### على Linux / macOS (باستخدام `make`)

```bash
make all
```

هذا يبني كل الأنظمة الخمسة دفعة واحدة داخل مجلد `dist/`.

أوامر مفردة متاحة أيضًا:
```bash
make windows     # Windows فقط
make linux       # Linux amd64 فقط
make macos-arm   # macOS Apple Silicon فقط
make clean       # تنظيف مجلد dist
make verify      # التحقق أن الملف الناتج لا يعتمد على مكتبات خارجية
```

### على Linux / macOS (بدون `make`، عبر bash مباشرة)

```bash
chmod +x build.sh
./build.sh
```

### على Windows (من CMD مباشرة، بدون WSL أو Git Bash)

```cmd
build.bat
```

جميع الطرق الثلاث تنتج نفس النتيجة: 5 ملفات تنفيذية داخل مجلد `dist/`.

---

## 4. لماذا الملف الناتج "نظيف" ولا يحتاج تثبيت أي شيء على جهاز المستخدم؟

استخدمنا 3 تقنيات أساسية في `Makefile` و`build.sh` لضمان ذلك:

1. **`CGO_ENABLED=0`**
   يوقف الربط بمكتبات C الخارجية (glibc وغيرها) ويجبر Go على بناء ملف **مرتبط ارتباطًا ثابتًا بالكامل (Fully Static Binary)**. هذا يعني أن الملف لا "يستدعي" أي مكتبة نظام خارجية وقت التشغيل.

2. **`-ldflags="-s -w"`**
   يحذف معلومات التصحيح (Debug symbols) وجدول الرموز (Symbol table) من الملف الناتج، ما يقلل الحجم بشكل ملحوظ (أحيانًا يقل الحجم للنصف) دون التأثير على الأداء.

3. **`-trimpath`**
   يزيل مسارات نظام الملفات الخاصة بجهازك (مثل `/home/username/...`) من الملف الناتج، لأسباب أمنية واحترافية (لا تريد أن يرى المشتري مساراتك الشخصية داخل الملف الثنائي).

### كيف تتحقق بنفسك أن الملف "نظيف"؟

على Linux/macOS:
```bash
file dist/fastsub-scanner-linux-amd64
# يجب أن تظهر كلمة "statically linked"

ldd dist/fastsub-scanner-linux-amd64
# يجب أن تظهر رسالة "not a dynamic executable"
# (هذا يعني: لا يعتمد على أي مكتبة خارجية على الإطلاق)
```

تم تنفيذ هذا الفحص فعليًا أثناء إعداد هذا الدليل، وكانت النتيجة:
```
dist/fastsub-scanner-linux-amd64: ELF 64-bit LSB executable, x86-64, statically linked, stripped
not a dynamic executable
```

بالنسبة لملفات Windows (`.exe`) وmacOS، أدوات `CGO_ENABLED=0` تضمن نفس السلوك (Static linking)، لكن أدوات الفحص (`ldd`, `otool -L`) تحتاج تشغيلها على نفس النظام المستهدف للتحقق البصري؛ التقنية نفسها مضمونة عبر متغيرات البناء بغض النظر عن نظام التشغيل الذي تبني منه.

---

## 5. تجربة فعلية قبل الشحن للمشتري

قبل تسليم أي ملف تنفيذي، شغّله فعليًا على نظامه المستهدف (أو عبر جهاز/VM يماثله) للتأكد:

```bash
echo "github.com" > test.txt
./fastsub-scanner-linux-amd64 -l test.txt
```

النتيجة المتوقعة (تم اختبارها فعليًا):
```
[*] تم تحميل 1 نطاق للفحص
[*] عدد الخيوط المتزامنة: 50 | المهلة: 5s

[Alive] github.com                          | 200 | https  | Server: github.com

[*] اكتمل الفحص في ...
[*] النتيجة: 1 فعّال / 1 إجمالي
```

على Windows، المستخدم يشغّل الملف مباشرة بالنقر المزدوج (لن يظهر مربع نص لأنه CLI — يفضّل توجيهه لتشغيله من `cmd.exe` أو `PowerShell`) أو عبر:
```powershell
.\fastsub-scanner-windows-amd64.exe -l domains.txt
```

---

## 6. نصيحة للبيع

عند تسليم المنتج للمشتري، أرفق:
- الملف التنفيذي المناسب لنظامه (لا ترسل الكود المصدري إلا إذا كانت الصفقة تشمله)
- ملف `README.md`
- ملف نطاقات تجريبي صغير (`sample-domains.txt`) ليجرب الأداة فورًا

هذا يقلل أسئلة الدعم الفني بعد البيع بشكل كبير.
