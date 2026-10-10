# EVAL FIXTURE — vulnerable on purpose. Do not copy.
from sqlalchemy import Column, Float, ForeignKey, Integer, String
from sqlalchemy.orm import declarative_base

Base = declarative_base()


class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    email = Column(String, unique=True, nullable=False)


class Account(Base):
    __tablename__ = "accounts"
    id = Column(Integer, primary_key=True)
    owner_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    balance = Column(Float, nullable=False, default=0)


class Topup(Base):
    __tablename__ = "topups"
    id = Column(Integer, primary_key=True)
    reference = Column(String, nullable=False)
    account_id = Column(Integer, ForeignKey("accounts.id"), nullable=False)
    amount = Column(Float, nullable=False)


class Coupon(Base):
    __tablename__ = "coupons"
    id = Column(Integer, primary_key=True)
    code = Column(String, unique=True, nullable=False)
    value = Column(Float, nullable=False)


class Redemption(Base):
    __tablename__ = "redemptions"
    id = Column(Integer, primary_key=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    coupon_id = Column(Integer, ForeignKey("coupons.id"), nullable=False)
