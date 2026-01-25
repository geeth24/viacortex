import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { cookies } from "next/headers";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  try {
    const cookieStore = await cookies();
    const session = cookieStore.get("session");

    if (!session?.value) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { id: domainId } = await params;

    const response = await fetch(
      `${process.env.BACKEND_URL}/api/domains/${domainId}`,
      {
        headers: {
          Authorization: `Bearer ${session.value}`,
          "Content-Type": "application/json",
        },
      },
    );

    if (!response.ok) {
      const errorResponse = response.clone();
      try {
        const errorData = await response.json();
        return NextResponse.json(errorData, { status: response.status });
      } catch {
        const errorText = await errorResponse.text();
        return NextResponse.json(
          { error: errorText || "Unknown error" },
          { status: response.status },
        );
      }
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching domain:", error);
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 },
    );
  }
}

export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  try {
    const cookieStore = await cookies();
    const session = cookieStore.get("session");

    if (!session?.value) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { id: domainId } = await params;
    const body = await request.json();

    const response = await fetch(
      `${process.env.BACKEND_URL}/api/domains/${domainId}`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${session.value}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      },
    );

    if (!response.ok) {
      const errorResponse = response.clone();
      try {
        const errorData = await response.json();
        return NextResponse.json(errorData, { status: response.status });
      } catch {
        const errorText = await errorResponse.text();
        return NextResponse.json(
          { error: errorText || "Unknown error" },
          { status: response.status },
        );
      }
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error updating domain:", error);
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 },
    );
  }
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> },
) {
  try {
    const cookieStore = await cookies();
    const session = cookieStore.get("session");

    if (!session?.value) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { id: domainId } = await params;

    const response = await fetch(
      `${process.env.BACKEND_URL}/api/domains/${domainId}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${session.value}`,
          "Content-Type": "application/json",
        },
      },
    );

    if (!response.ok) {
      const errorResponse = response.clone();
      try {
        const errorData = await response.json();
        return NextResponse.json(errorData, { status: response.status });
      } catch {
        const errorText = await errorResponse.text();
        return NextResponse.json(
          { error: errorText || "Unknown error" },
          { status: response.status },
        );
      }
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Error deleting domain:", error);
    return NextResponse.json(
      { error: "Internal Server Error" },
      { status: 500 },
    );
  }
}
