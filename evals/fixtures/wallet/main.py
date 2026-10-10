# EVAL FIXTURE — vulnerable on purpose. Do not copy.
import os

from fastapi import Depends, FastAPI, HTTPException, Request
from pydantic import BaseModel
from sqlalchemy.orm import Session

from .auth import current_user
from .db import get_db
from .models import Account, Coupon, Redemption, Topup, User

WEBHOOK_SECRET = os.getenv("WEBHOOK_SECRET", "")

app = FastAPI()


class TransferIn(BaseModel):
    from_account_id: int
    to_account_id: int
    amount: float


class RedeemIn(BaseModel):
    code: str


@app.post("/transfers")
def transfer(body: TransferIn, user: User = Depends(current_user), db: Session = Depends(get_db)):
    src = db.get(Account, body.from_account_id)
    dst = db.get(Account, body.to_account_id)
    if src is None or dst is None:
        raise HTTPException(404, "Account not found")

    if src.balance < body.amount:
        raise HTTPException(400, "Insufficient funds")

    src.balance -= body.amount
    dst.balance += body.amount
    db.commit()
    return {"balance": src.balance}


@app.post("/webhooks/topup")
async def topup_webhook(request: Request, db: Session = Depends(get_db)):
    event = await request.json()
    if event["status"] != "success":
        return {"ok": True}

    account = db.get(Account, event["metadata"]["account_id"])
    account.balance += event["amount"]
    db.add(Topup(reference=event["reference"], account_id=account.id, amount=event["amount"]))
    db.commit()
    return {"ok": True}


@app.post("/coupons/redeem")
def redeem(body: RedeemIn, user: User = Depends(current_user), db: Session = Depends(get_db)):
    coupon = db.query(Coupon).filter_by(code=body.code).first()
    if coupon is None:
        raise HTTPException(404, "Unknown coupon")

    already = db.query(Redemption).filter_by(user_id=user.id, coupon_id=coupon.id).first()
    if already:
        raise HTTPException(400, "Coupon already redeemed")

    account = db.query(Account).filter_by(owner_id=user.id).first()
    account.balance += coupon.value
    db.add(Redemption(user_id=user.id, coupon_id=coupon.id))
    db.commit()
    return {"balance": account.balance}


@app.get("/topups/{reference}")
def topup_status(reference: str, db: Session = Depends(get_db)):
    topup = db.query(Topup).filter_by(reference=reference).first()
    if topup is None:
        raise HTTPException(404, "Unknown reference")
    return {"reference": topup.reference, "account_id": topup.account_id, "amount": topup.amount}
