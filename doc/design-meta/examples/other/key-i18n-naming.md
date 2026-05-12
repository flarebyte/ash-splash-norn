# Recommended ARB key naming conventions

* **Start with a stable domain/feature prefix**

  * `authLoginTitle`, `checkoutPaymentErrorCardDeclined`
  * 👉 Makes large files searchable and scalable

* **Use camelCase (Dart-friendly)**

  * 👉 Matches generated `AppLocalizations` API

* **Structure loosely as:**

  * `feature + scope + semanticMeaning`
  * 👉 Not rigid `<feature><screen><element><intent>`, but close enough

* **End with a role/intent when useful**

  * `Title`, `Label`, `Button`, `Error`, `DialogBody`
  * 👉 Clarifies usage in UI

* **Prefer semantic meaning over UI position**

  * `DiscardChangesDialogBody` ✅
  * `TopText` ❌

* **Keep keys stable (don’t encode wording)**

  * `checkoutPayNowButton` ✅
  * `payNowInsteadOfCheckoutButton` ❌

* **Avoid generic or reused keys**

  * `saveButton` ❌
  * `profileEditSaveButton` ✅

* **Use prefixes for grouping (not multiple ARB files)**

  * `auth...`, `checkout...`, `settings...`
  * 👉 Works with Flutter’s single-file-per-locale model

---

# 🧠 Practical reasons behind these rules

* **Scalability**

  * Hundreds/thousands of keys stay navigable

* **Searchability**

  * Easy to find by feature (`auth`, `checkout`, etc.)

* **Stability**

  * Keys survive UI and copy changes

* **Readability in code**

  * `AppLocalizations.of(context)!.profileEditTitle` is self-explanatory

* **Avoid collisions**

  * No ambiguity like multiple `title` or `submit`

* **Works with Flutter tooling**

  * Clean generated Dart API

* **Flexible for real-world complexity**

  * Doesn’t break when concepts don’t fit strict schemas

---

# ❌ Why too strict naming conventions don’t work well

* **Reality doesn’t fit the schema**

  * Not everything is cleanly `<feature><screen><element><intent>`
  * Some strings belong to flows, states, or reusable components
  * 👉 You end up forcing awkward or fake structure

---

* **Leads to artificial keys**

  * You invent parts just to satisfy the format
  * 👉 Keys become verbose and less meaningful

---

* **Breaks easily when the app evolves**

  * Screens get renamed, merged, or removed
  * 👉 Either keys become outdated or you rename them constantly

---

* **Hurts readability**

  * Over-structured keys are harder to scan in code
  * 👉 Developers read these far more than machines parse them

---

* **Low practical payoff**

  * You rarely need to reconstruct structure from keys
  * 👉 Parsing keys brings little real benefit

---

* **Encodes the wrong thing**

  * Keys should represent **meaning**, not exact UI structure
  * 👉 UI structure changes more often than meaning

---

# ✅ Better approach

* Keep keys:

  * **loosely structured**
  * **human-readable**
  * **stable**

* Put strict structure (if needed) in:

  * ARB metadata (`@key`)
  * or external config
