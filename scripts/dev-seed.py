#!/usr/bin/env python3
"""Seed the LOCAL dev API (scripts/dev-local.sh) with a ready-to-ride operator. Refuses any non-local URL.

Creates: platform admin, operator "Lanka Test Cabs" (slug lanka-test) covering all of Sri Lanka, car + tuk
categories and fares, an approved vehicle, and an approved driver with documents. Test-only credentials below.
Passengers sign up from the passenger app themselves (the 2026-10-11 device run used rider@taxios.test / taxios-dev-rider-1).
"""
import json, os, subprocess, sys, urllib.request, urllib.error

API = os.environ.get("TAXIOS_API_URL", "http://localhost:8080")
assert API.startswith(("http://localhost", "http://127.0.0.1")), "dev-seed only runs against a local API"

# Test-only values for the throwaway local database.
ADMIN = ("admin@taxios.test", "taxios-dev-admin-1")
OWNER = ("owner@taxios.test", "taxios-dev-owner-1")
DRIVER = ("driver@taxios.test", "taxios-dev-driver-1", "Nimal Perera", "0771234567")
SLUG = "lanka-test"
SRI_LANKA = {"type": "Polygon", "coordinates": [[[79.4, 5.8], [82.0, 5.8], [82.0, 9.95], [79.4, 9.95], [79.4, 5.8]]]}

def call(method, path, token=None, body=None, ok=(200, 201, 204)):
    req = urllib.request.Request(API + path, method=method, data=json.dumps(body).encode() if body is not None else None)
    req.add_header("Content-Type", "application/json")
    if token: req.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(req) as r:
            raw = r.read(); status = r.status
    except urllib.error.HTTPError as e:
        raw = e.read(); status = e.code
    data = json.loads(raw) if raw else {}
    if status not in ok: raise SystemExit(f"{method} {path} -> {status} {data}")
    return status, data

def login(email, pw):
    return call("POST", "/v1/auth/login", body={"email": email, "password": pw, "client": "mobile"})[1]["token"]

# Admin (CLI bootstrap, same as production).
env = dict(os.environ, TAXIOS_ADMIN_PASSWORD=ADMIN[1],
           MIGRATE_DATABASE_URL="postgres://taxios_owner:taxios_dev_owner@127.0.0.1:5433/taxios?sslmode=disable")
subprocess.run([".dev/taxios-api", "create-platform-admin", ADMIN[0], "Dev Admin"], env=env, check=True, capture_output=True)
admin = login(*ADMIN)

st, pub = call("GET", f"/v1/public/tenants/{SLUG}", ok=(200, 404))
if st == 200:
    print(f"already seeded: operator code '{SLUG}'"); sys.exit(0)

_, r = call("POST", "/v1/platform/tenants", admin, {"slug": SLUG, "name": "Lanka Test Cabs", "businessType": "taxi", "ownerEmail": OWNER[0]})
tid = r["tenant"]["id"]; base = f"/v1/tenants/{tid}"
owner = call("POST", "/v1/auth/accept-invite", body={"token": r["ownerInvitation"]["token"], "fullName": "Owner Test", "password": OWNER[1]})[1]["token"]
call("PATCH", base, owner, {"legalName": "Lanka Test Cabs (Pvt) Ltd", "registrationNo": "PV 00001", "contactEmail": "ops@lankatest.test", "contactPhone": "0112345678"})
call("PATCH", f"{base}/branding", owner, {"color": "#1d4ed8", "tagline": "Local test operator", "supportPhone": "0112345678"})
zone = call("POST", f"{base}/zones", owner, {"name": "Sri Lanka", "area": SRI_LANKA})[1]["zone"]["id"]
car = call("POST", f"{base}/vehicle-categories", owner, {"code": "car", "name": "Car", "seats": 4})[1]["category"]["id"]
tuk = call("POST", f"{base}/vehicle-categories", owner, {"code": "tuk", "name": "Tuk", "seats": 3})[1]["category"]["id"]
for cat, base_fare, per_km in ((car, 10000, 8500), (tuk, 6000, 6000)):
    call("POST", f"{base}/fare-rules", owner, {"zoneId": zone, "categoryId": cat, "baseMinor": base_fare, "perKmMinor": per_km, "perMinuteMinor": 500,
         "minimumMinor": 25000, "bookingFeeMinor": 2000, "waitingPerMinMinor": 300, "freeWaitingMinutes": 3, "roundToMinor": 100})
call("POST", f"{base}/onboarding/complete", owner)

veh = call("POST", f"{base}/vehicles", owner, {"categoryId": car, "plate": "CAB-1234", "make": "Toyota", "model": "Axio", "colour": "White"})[1]["vehicle"]["id"]
call("POST", f"{base}/vehicles/{veh}/status", owner, {"status": "approved"})
_, r = call("POST", f"{base}/drivers", owner, {"fullName": DRIVER[2], "phone": DRIVER[3], "email": DRIVER[0]})
did = r["driver"]["id"]
drv = call("POST", "/v1/auth/accept-invite", body={"token": r["invitation"]["token"], "fullName": DRIVER[2], "password": DRIVER[1]})[1]["token"]
for doc in ("driving_licence", "national_id", "vehicle_insurance"):
    call("POST", f"{base}/driver/documents", drv, {"docType": doc, "number": doc.upper()[:3] + "-0001", "expiresOn": "2030-12-31"})
for d in call("GET", f"{base}/drivers/{did}", owner)[1]["documents"]:
    call("POST", f"{base}/drivers/{did}/documents/{d['id']}/review", owner, {"decision": "approved"})
call("PUT", f"{base}/drivers/{did}/vehicle", owner, {"vehicleId": veh})
call("POST", f"{base}/drivers/{did}/status", owner, {"status": "approved"})
print(f"seeded: operator code '{SLUG}' (all of Sri Lanka), driver {DRIVER[0]} approved with vehicle CAB-1234")
